package conn

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
)

type PeerConn struct {
	conn        net.Conn
	localPeerID string
	remotePeer  *torrent.Peer
	infohash    string
}

func EstablishConnection(localPeerID string, rp *torrent.Peer, infohash string) (*PeerConn, error) {
	pc := &PeerConn{
		localPeerID: localPeerID,
		remotePeer:  rp,
		infohash:    infohash,
	}
	rpid, conn, err := pc.performHandshake()
	if err != nil {
		return nil, err
	}
	pc.remotePeer.PeerID = rpid
	pc.conn = conn

	return pc, nil
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
		return "", nil, err
	}
	// TODO: Don't close the conn now
	defer conn.Close()

	n, err := conn.Write(msg.serialize())
	if n != len(msg.serialize()) || err != nil {
		return "", nil, fmt.Errorf("error writing msg: %v", err)
	}

	respData := make([]byte, 68)
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		return "", nil, err
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

func (pc *PeerConn) Close() error {
	return pc.conn.Close()
}

func (pc *PeerConn) RemotePeerID() string {
	return pc.remotePeer.PeerID
}
