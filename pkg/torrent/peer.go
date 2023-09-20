package torrent

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Peer struct {
	AddrIPV4 string
	Port     uint16

	PeerID string
}

func (p *Peer) PerformHandshake(infohash string) (string, error) {
	msg := &PeerHandshakeMsg{
		protocolLen: 19,
		protocol:    "BitTorrent protocol",
		reserved:    make([]byte, 8),
		infohash:    []byte(infohash),
		peerId:      peerID, // own peer id, not peer's
	}

	conn, err := net.Dial("tcp", p.AddrIPV4+":"+strconv.Itoa(int(p.Port)))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	n, err := conn.Write(msg.serialize())
	if n != len(msg.serialize()) || err != nil {
		return "", fmt.Errorf("error writing msg: %v", err)
	}

	respData := make([]byte, 68)
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		return "", err
	}

	hsResp, err := deserializePeerHandshakeMsg(respData)
	if err != nil {
		return "", err
	}
	if !hsResp.validate() {
		return "", fmt.Errorf("invalid handshake response")
	}

	return hsResp.peerId, nil

}

func peerFromBytes(b []byte) (*Peer, error) {
	if len(b) != 6 {
		return nil, fmt.Errorf("too few bytes provided to construct peer")
	}
	return &Peer{
		AddrIPV4: fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3]),
		Port:     binary.BigEndian.Uint16(b[4:6]),
	}, nil
}
