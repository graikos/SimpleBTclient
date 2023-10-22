package services

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/conn"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/log"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
)

type DownloadFileService interface {
	DownloadFile(string, string) error
}

type downloadFileServiceImpl struct {
	logger log.Logger
}

func NewDownloadFileService() DownloadFileService {
	return &downloadFileServiceImpl{log.NewLogger(log.NORMAL)}
}

func (df *downloadFileServiceImpl) DownloadFile(torrentFile, filepath string) error {
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

	// pieceLen, err := t.PieceLength()
	// if err != nil {
	// 	return err
	// }
	// tLen, err := t.Length()
	// if err != nil {
	// 	return err
	// }

	pieces, err := t.Pieces()
	if err != nil {
		return err
	}

	pieceStorage := make([]*bytes.Buffer, len(pieces))

	// place all the indexes as tasks in a queue
	pieceQueue := make(chan int, len(pieces))
	df.logger.Info("Torrent no of pieces:", len(pieces))

	for i := 0; i < len(pieces); i++ {
		pieceQueue <- i
		df.logger.Debug("Initially added task with idx to queue", i)
		pieceStorage[i] = new(bytes.Buffer)
	}

	successChan := make(chan struct{})

	go func() {
		for pidx := range pieceQueue {
			pidx := pidx
			go func() {

				selectedPeer := resp.Peers[rand.Intn(len(resp.Peers))]
				df.logger.Debug(selectedPeer)
				peerConn, err := conn.EstablishConnection(torrent.LocalPeerID, selectedPeer, t, Logger)
				df.logger.Debug("worker established one connection for idx", pidx, "with peer:", selectedPeer)
				if err != nil {
					// if error occurs put back in queue
					pieceQueue <- pidx
					df.logger.Debug(err)
					return
				}
				defer func() {
					if err := peerConn.Close(); err != nil {
						df.logger.Debug(err)
					}
				}()
				err = peerConn.AskForPiece(pidx, pieceStorage[pidx])
				if err != nil {
					// if error occurs put back in queue
					pieceQueue <- pidx
					df.logger.Debug(err)
					return
				}
				df.logger.Info("Piece with idx", pidx, "downloaded")

				successChan <- struct{}{}

			}()
		}
	}()

	counter := len(pieces)

	for {
		<-successChan
		if counter--; counter == 0 {
			close(pieceQueue)
			break
		}
	}

	close(successChan)

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return err
	}

	for _, p := range pieceStorage {
		io.Copy(f, p)
	}

	return nil

}
