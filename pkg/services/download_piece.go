package services

import (
	"fmt"
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

	for _, remotePeer := range resp.Peers {
		// // randomly select one peer
		// remotePeer := resp.Peers[0]

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

	return peerConn.AskForPiece(idx, filepath)
}
