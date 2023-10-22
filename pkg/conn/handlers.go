package conn

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/util"
)

const blockSize int = 16 * 1024

type block struct {
	idx    int
	begin  int
	length int
}

func (pc *PeerConn) produceInterested(e *event) error {

	// setting current piece index
	currentIdx := int(binary.BigEndian.Uint32(e.payload[0:4]))

	var curPieceLen int

	pieceLen, err := pc.torrent.PieceLength()
	if err != nil {
		return err
	}
	tLen, err := pc.torrent.Length()
	if err != nil {
		return err
	}

	curPieceLen = util.GetLengthForIdx(tLen, pieceLen, currentIdx)

	pc.currentPiece = torrent.NewPiece(curPieceLen, pc.storage, currentIdx)

	return pc.write(newPeerMessage(interested, []byte{}))
}

func (pc *PeerConn) produceRequest(e *event) {

	curPieceLen := pc.currentPiece.Length()

	// split piece in blocks
	noOfBlocks := int((curPieceLen + blockSize - 1) / blockSize)

	q := make(chan struct{}, pipelineRequestsLimit)
	errChan := make(chan error, noOfBlocks)
	wg := new(sync.WaitGroup)

	for i := 0; i < noOfBlocks; i++ {

		begin := i * blockSize

		// determine block size and handle edge case of uneven division and piece being last
		l := blockSize
		if i == noOfBlocks-1 {
			l = curPieceLen % blockSize
			if l == 0 {
				l = blockSize
			}
		}

		// buffered channel to have a maximum of pipelineRequestsLimit
		// workers running
		q <- struct{}{}
		wg.Add(1)
		go pc.requestBlock(q, wg, &block{
			idx:    pc.currentPiece.Index(),
			begin:  begin,
			length: l,
		}, errChan)
	}

	go func() {
		wg.Wait()
		pc.logger.Debug("Done waiting in requests routine")
		close(errChan)
		close(q)
	}()

	// execution will block here while go routines are still active
	// and until all errors are read
	for err := range errChan {
		if err != nil {
			// pipe to main PeerConnection error channel for
			// handling in the main event handling routine
			// main error channel buffer size does not matter,
			// even if this blocks, eventually main handling routine will
			// consume the errors
			pc.errChan <- err
		}
	}
}

func (pc *PeerConn) requestBlock(q <-chan struct{}, wg *sync.WaitGroup, b *block, errChan chan<- error) {

	pc.logger.Debug("Making block of index:", b.idx, "begin:", b.begin, "length:", b.length)

	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(b.idx))
	binary.BigEndian.PutUint32(payload[4:8], uint32(b.begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(b.length))

	errChan <- pc.write(newPeerMessage(request, payload))

	// task done, release worker
	wg.Done()
	<-q
}

func (pc *PeerConn) handlePiece(e *event) {

	// unmarshal payload
	pieceIdxReceived := int(binary.BigEndian.Uint32(e.payload[0:4]))
	if pieceIdxReceived != pc.currentPiece.Index() {
		pc.errChan <- fmt.Errorf("piece index mismatch")
		return
	}
	begin := int(binary.BigEndian.Uint32(e.payload[4:8]))
	blockData := e.payload[8:len(e.payload)]

	// no race conditions since different locations are assumed
	pc.currentPiece.WriteBlock(begin, blockData)

	// check if complete
	if !(pc.currentPiece.IsComplete()) {
		return
	}

	hashes, err := pc.torrent.Pieces()
	if err != nil {
		pc.errChan <- err
		return
	}

	// verify integrity of the piece received
	if !pc.currentPiece.Verify(hashes[pc.currentPiece.Index()]) {
		pc.errChan <- fmt.Errorf("actual and expected piece hash mismatch")
		return
	}

	if err := pc.currentPiece.Commit(); err != nil {
		// TODO: add channel to event handler
		pc.errChan <- err
		return
	}

	// NOTE: maybe reset piece field if needed, in this case
	// producing request messages will set it to the new piece

	// signal end of piece download
	close(e.signal)
}
