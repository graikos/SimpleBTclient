package services

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/conn"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/log"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
)

var Logger log.Logger

type DownloadPieceService interface {
	DownloadPiece(string, string, int) error
}

type downloadPieceServiceImpl struct {
}

func NewDownloadPieceService() DownloadPieceService {
	return &downloadPieceServiceImpl{}
}

func (dps *downloadPieceServiceImpl) DownloadPiece(filepath string, torrentFile string, idx int) error {
	t, err := torrent.NewSingleTorrentFromFile(torrentFile)
	if err != nil {
		return err
	}
	tracker := torrent.NewTracker(t)

	resp, err := tracker.AskForPeers()
	if err != nil {
		return err
	}

	if len(resp.Peers) == 0 {
		return fmt.Errorf("no peers found")
	}

	var peerConn *conn.PeerConn

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return err
	}

	for _, remotePeer := range resp.Peers {

		peerConn, err = conn.EstablishConnection(torrent.LocalPeerID, remotePeer, t, Logger)
		if err != nil {
			if strings.Contains(err.Error(), "reading handshake") {
				continue
			}
			return err
		}
		Logger.Info("Established connection with peer: ", remotePeer.AddrIPV4, remotePeer.Port)
		break
	}

	defer peerConn.Close()

	return peerConn.AskForPiece(idx, f)
}
