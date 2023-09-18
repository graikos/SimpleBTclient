package torrent

import (
	"encoding/hex"
	"fmt"
	"testing"
)

type mockTorrent struct {
}

func (m *mockTorrent) Length() (int, error) {
	return 32768, nil
}

func (m *mockTorrent) InfoHash() ([]byte, error) {
	return hex.DecodeString("d69f91e6b2ae4c542468d1073a71d4ea13879a7f")
}

func (m *mockTorrent) Announce() string {
	return "http://bittorrent-test-tracker.codecrafters.io/announce"
}

func newTestTorrent() Torrent {
	return &mockTorrent{}
}

func TestTrackerAskForPeers(t *testing.T) {
	tracker := NewTracker(newTestTorrent())
	trackerResponse, err := tracker.AskForPeers()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(trackerResponse)
	for _, peer := range trackerResponse.Peers {
		fmt.Println(peer)
	}
}
