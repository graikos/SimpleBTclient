package torrent

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bencode"
)

var ErrInvalidTrackerResponseFormat = errors.New("invalid tracker response format")

type Tracker struct {
	torrent Torrent
}

var peerID string = "00112233445566778899"

func NewTracker(torrent Torrent) *Tracker {
	return &Tracker{
		torrent: torrent,
	}
}

func (t *Tracker) AskForPeers() (*TrackerResponse, error) {
	params := url.Values{}

	infohash, err := t.torrent.InfoHash()
	if err != nil {
		return nil, err
	}

	params.Add("info_hash", string(infohash))
	params.Add("peer_id", peerID) // hardcoding this one
	params.Add("port", "6881")    // won't be used
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")

	l, err := t.torrent.Length()
	if err != nil {
		return nil, err
	}
	params.Add("left", strconv.Itoa(l))
	params.Add("compact", "1")

	reqUrl := t.torrent.Announce() + "?" + params.Encode()

	response, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println("body is")
	// fmt.Println(string(body))

	decodedResp, err := bencode.DecodeBencode(string(body))
	if err != nil {
		return nil, err
	}
	respDict, ok := decodedResp.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidTrackerResponseFormat
	}

	// each peer holds 6 bytes in the response
	peersProvidedStr, ok := respDict["peers"].(string)
	if !ok {
		return nil, ErrInvalidTrackerResponseFormat
	}

	peersProvided := []byte(peersProvidedStr)

	// check if peers string is a multiple of 6
	if len(peersProvided)%6 != 0 {
		return nil, fmt.Errorf("invalid peers field")
	}

	peers := make([]*Peer, len(peersProvided)/6)

	for i := range peers {
		// create peer for every 6 bytes
		peers[i], err = peerFromBytes(peersProvided[i*6 : (i+1)*6])
		if err != nil {
			return nil, err
		}
	}

	return &TrackerResponse{
		Interval: respDict["interval"].(int),
		Peers:    peers,
	}, nil
}
