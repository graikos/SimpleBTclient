package conn

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/log"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/util/fsm"
)

const (
	eventQueueSize        = 15
	pipelineRequestsLimit = 5
)

type PeerConn struct {
	mu   sync.Mutex
	conn net.Conn

	localPeerID string
	remotePeer  *torrent.Peer
	infohash    string

	torrent torrent.Torrent

	fsm *fsm.FSM

	hasBitfield chan struct{}

	currentPiece torrent.Piece
	storage      io.Writer

	eventQueue chan *event
	errChan    chan error

	logger log.Logger
}

type event struct {
	name    string
	payload []byte
	signal  chan error
}

const (
	waitingForBitfield int = iota
	haveBitfield
	waitingForUnchoke
	waitingForPiece
	waitingForInit
	sentInterested
	receivedUnchoke
	receivingPieces
)

func EstablishConnection(localPeerID string, rp *torrent.Peer, t torrent.Torrent, logger log.Logger) (*PeerConn, error) {
	pc := &PeerConn{
		localPeerID: localPeerID,
		remotePeer:  rp,
		torrent:     t,
		logger:      logger,
	}
	ih, err := t.InfoHash()
	if err != nil {
		return nil, err
	}
	pc.infohash = string(ih)

	rpid, conn, err := pc.performHandshake()
	if err != nil {
		return nil, err
	}
	pc.remotePeer.PeerID = rpid
	pc.conn = conn
	pc.initFSM()

	pc.hasBitfield = make(chan struct{})

	// initialize msg queue
	pc.eventQueue = make(chan *event, eventQueueSize)
	// initialize error channel
	// will be used for handling the errors caused in event handlers
	// size will be 5 just so execution of main handleEvent doesn't block
	// and the select can select the error channel in the next poll
	pc.errChan = make(chan error, 5)

	// start listening to incoming messages
	go pc.listen()

	// start handling events
	go pc.handleEventQueue()

	return pc, nil
}

// AskForPiece will initiate a peer message exchange to download the piece specified by idx.
// Since the response messages do not identify a piece uniquely, only one piece can be downloaded at a time.
func (pc *PeerConn) AskForPiece(idx int, writer io.Writer) error {

	pc.logger.Debug("Started AskForPiece routine")

	// wait for bitfield message
	// this is optional in the bittorrent protocol but required in the codecrafters outline
	<-pc.hasBitfield

	// initialize the piece storage
	// each call to AskForPiece will renew this as expected
	pc.storage = writer

	pc.logger.Debug("Passed hasBitfield barrier in AskForPiece")

	// add an event that sends interested message
	// if the FSM is at a state where a new transfer can begin, the current idx will be set
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(idx))

	s := make(chan error)
	pc.eventQueue <- &event{
		name:    "initiated",
		payload: buf,
		signal:  s,
	}

	pc.logger.Debug("Just placed initiated event")

	// Wait for download to end and receive error
	err, ok := <-s
	// if channel closed, piece finished downloading
	if !ok || err == nil {
		pc.logger.Debug("AskPiece detected closing channel with err", err)
		return nil
	} else {
		return err
	}
}

func (pc *PeerConn) performHandshake() (string, net.Conn, error) {
	msg := &PeerHandshakeMsg{
		protocolLen: 19,
		protocol:    "BitTorrent protocol",
		reserved:    make([]byte, 8),
		infohash:    []byte(pc.infohash),
		peerId:      pc.localPeerID, // own peer id, not peer's
	}

	conn, err := net.Dial("tcp", pc.remotePeer.AddrIPV4+":"+strconv.Itoa(int(pc.remotePeer.Port)))
	if err != nil {
		return "", nil, fmt.Errorf("dialing: %v", err)
	}

	n, err := conn.Write(msg.serialize())
	if n != len(msg.serialize()) || err != nil {
		return "", nil, fmt.Errorf("error writing msg: %v", err)
	}

	respData := make([]byte, 68)
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		return "", nil, fmt.Errorf("reading handshake response: %v", err)
	}

	hsResp, err := deserializePeerHandshakeMsg(respData)
	if err != nil {
		return "", nil, err
	}
	if !hsResp.validate() {
		return "", nil, fmt.Errorf("invalid handshake response")
	}

	return hsResp.peerId, conn, nil
}

func (pc *PeerConn) listen() error {
	reader := bufio.NewReader(pc.conn)
	pc.logger.Debug("Listening on connection...")
	for {
		pc.logger.Debug("Listening loop...")
		// read the length prefix

		lenPrefix := make([]byte, 4)
		_, err := io.ReadFull(reader, lenPrefix)
		if err != nil {
			return fmt.Errorf("reading length prefix: %v", err)
		}

		msgLen := binary.BigEndian.Uint32(lenPrefix)
		pc.logger.Debug("Received message of length", msgLen)

		// allocate buffer of size msgLen to receive msg
		msgBuf := make([]byte, msgLen)

		pc.mu.Lock()
		pc.logger.Debug("just locked read")

		_, err = io.ReadFull(reader, msgBuf)
		if err != nil {
			return fmt.Errorf("reading message: %v", err)
		}

		pc.mu.Unlock()
		pc.logger.Debug("just unlocked read")

		if len(msgBuf) < 100 {
			pc.logger.Debug("read full message:", msgBuf)
		} else {
			pc.logger.Debug("read full message (trunc):", msgBuf[:100])
		}

		if len(msgBuf) > 0 {
			pc.eventQueue <- &event{
				name:    msgTypeToString[peerMsgType(msgBuf[0])],
				payload: msgBuf[1:msgLen],
			}
			pc.logger.Debug("just placed event in queue")
		}

	}
}

func (pc *PeerConn) handleEventQueue() error {
	var currentsig chan error
	for {
		select {
		case e, ok := <-pc.eventQueue:

			if !ok {
				// closed channel, means end the handling
				pc.logger.Debug("Event queue channel closed")
				return nil
			}

			pc.logger.Debug("Handler just got event with name:", e.name, "and payload len:", len(e.payload))

			fsmOutMsg, ok := pc.fsm.ApplyTransition(e.name)
			if !ok {
				pc.logger.Debug("Ignoring msg: %v\n", e)
			}

			pc.logger.Debug("FSM out message is:", fsmOutMsg)

			switch fsmOutMsg {
			case "have_bitfield":
				// received bitfield, block until exchange is initiated
				pc.hasBitfield <- struct{}{}

			case "interested":
				pc.logger.Debug("in interested case")
				// AskPiece creates the signal channel and waits
				currentsig = e.signal
				err := pc.produceInterested(e)
				if err != nil {
					pc.logger.Debug("Handle interested error: ", err)
					pc.errChan <- err
				}

			case "request":
				go pc.produceRequest(e)

			case "save_piece":
				// when needed, the signal channel that AskPiece is
				// waiting on will be passed to the handler
				e.signal = currentsig
				go pc.handlePiece(e)

			}

		case err := <-pc.errChan:
			pc.logger.Debug("Got error in handler routine:", err)
			currentsig <- err
		}
	}
}

// initFSM initializes the Finite State Machine that will keep track
// of the state and the response for each event according to it. This is
// a concise way to handle all different cases of receiving events asynchronously.
func (pc *PeerConn) initFSM() {
	m := make(map[fsm.TransitionInput]fsm.TransitionOutput)

	m[fsm.TransitionInput{OldState: waitingForBitfield, InMsg: "bitfield"}] =
		fsm.TransitionOutput{NewState: haveBitfield, OutMsg: "have_bitfield"}

	m[fsm.TransitionInput{OldState: haveBitfield, InMsg: "initiated"}] =
		fsm.TransitionOutput{NewState: sentInterested, OutMsg: "interested"}

	m[fsm.TransitionInput{OldState: sentInterested, InMsg: "unchoke"}] =
		fsm.TransitionOutput{NewState: receivedUnchoke, OutMsg: "request"}

	m[fsm.TransitionInput{OldState: receivedUnchoke, InMsg: "piece"}] =
		fsm.TransitionOutput{NewState: receivingPieces, OutMsg: "save_piece"}

	m[fsm.TransitionInput{OldState: receivingPieces, InMsg: "piece"}] =
		fsm.TransitionOutput{NewState: receivingPieces, OutMsg: "save_piece"}

	// fsm will be initialized with a state of waiting for the first bitfield message
	pc.fsm = fsm.NewFSM(m, waitingForBitfield)
}

func (pc *PeerConn) write(msg *peerMessage) error {

	pc.logger.Debug("received to write msg:", msg)

	// will hold msg length (4), message type (1) and payload (var)
	buf := make([]byte, len(msg.payload)+5)
	n := 0

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(msg.payload)+1))

	pc.logger.Debug("Writing msg of len: ", len(msg.payload)+1, "and lenbuf:", lenBuf)
	pc.mu.Lock()
	n, err := pc.conn.Write(lenBuf)
	if err != nil || n != 4 {
		return fmt.Errorf("writing msg len to conn: %v", err)
	}

	n = 0

	// copy message type to buffer
	buf[0] = byte(msg.msgType)
	// copy rest of message payload
	copy(buf[1:], msg.payload)

	pc.logger.Debug("Writing msg buf:", buf)

	writer := bufio.NewWriter(pc.conn)

	for n < len(buf) {
		t, err := writer.Write(buf)
		if err != nil {
			return fmt.Errorf("writing msg to buffered conn writer: %v", err)
		}
		n += t
	}

	err = writer.Flush()
	pc.mu.Unlock()
	if err != nil {
		return fmt.Errorf("flushing msg to conn: %v", err)
	}

	pc.logger.Debug("Message writing ended successfully")

	return nil
}

func (pc *PeerConn) Close() error {
	close(pc.hasBitfield)
	close(pc.eventQueue)
	// will stop handling routine
	close(pc.errChan)
	// will stop listening routine
	return pc.conn.Close()
}

func (pc *PeerConn) RemotePeerID() string {
	return pc.remotePeer.PeerID
}
