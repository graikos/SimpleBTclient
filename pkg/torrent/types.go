package torrent

import (
	"bytes"
	"fmt"
)

type TrackerResponse struct {
	Interval int
	Peers    []*Peer
}

type PeerHandshakeMsg struct {
	protocolLen byte
	protocol    string
	reserved    []byte
	infohash    []byte
	peerId      string
}

func (p *PeerHandshakeMsg) serialize() []byte {
	// known before hand, length is 20(protocol) + 8 reserved + 40 infohash
	// and peerId
	res := make([]byte, 68)
	res[0] = p.protocolLen
	copy(res[1:20], []byte(p.protocol))
	// default value is 0, so 8 reserved left as they are
	copy(res[20:28], p.reserved)
	copy(res[28:48], p.infohash)
	copy(res[48:68], []byte(p.peerId))
	return res
}

func deserializePeerHandshakeMsg(data []byte) (*PeerHandshakeMsg, error) {
	if len(data) != 68 {
		return nil, fmt.Errorf("invalid handshake message length")
	}
	return &PeerHandshakeMsg{
		protocolLen: data[0],
		protocol:    string(data[1:20]),
		reserved:    data[20:28],
		infohash:    data[28:48],
		peerId:      string(data[48:68]),
	}, nil
}

// validate will check against the given message structure
func (p *PeerHandshakeMsg) validate() bool {
	return (p.protocolLen == 19) || (p.protocol == "BitTorrent protocol") ||
		(bytes.Compare(p.reserved, make([]byte, 8)) == 0)
}
