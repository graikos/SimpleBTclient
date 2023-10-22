package conn

type peerMsgType int

const (
	choke peerMsgType = iota
	unchoke
	interested
	notInterested
	have
	bitfield
	request
	piece
	cancel
)

var msgTypeToString = map[peerMsgType]string{
	choke:         "choke",
	unchoke:       "unchoke",
	interested:    "interested",
	notInterested: "notInterested",
	have:          "have",
	bitfield:      "bitfield",
	request:       "request",
	piece:         "piece",
	cancel:        "cancel",
}

type peerMessage struct {
	msgType peerMsgType
	payload []byte
}

func newPeerMessage(t peerMsgType, payload []byte) *peerMessage {
	return &peerMessage{
		msgType: t,
		payload: payload,
	}
}
