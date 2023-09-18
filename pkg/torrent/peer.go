package torrent

import (
	"encoding/binary"
	"fmt"
)

type Peer struct {
	AddrIPV4 string
	Port     uint16
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
