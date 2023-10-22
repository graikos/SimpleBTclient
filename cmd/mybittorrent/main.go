package main

import (
	// Uncomment this line to pass the first stage

	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/conn"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/log"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/services"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {

	logger := log.NewLogger(log.NORMAL)
	services.Logger = logger

	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		jsonOutput, err := bencode.DecodeBencodeToJSON(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(jsonOutput)
	case "info":

		torrentPath := os.Args[2]

		t, err := torrent.NewSingleTorrentFromFile(torrentPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Tracker URL:", t.Announce())
		l, err := t.Length()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Length:", l)

		hash, err := t.InfoHash()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Info Hash: %x\n", hash)

		pl, err := t.PieceLength()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Piece Length:", pl)

		pcs, err := t.Pieces()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Piece Hashes:")
		for _, pieceHash := range pcs {
			fmt.Printf("%x\n", pieceHash)
		}

	case "peers":

		torrentPath := os.Args[2]

		t, err := torrent.NewSingleTorrentFromFile(torrentPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tracker := torrent.NewTracker(t)
		resp, err := tracker.AskForPeers()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, peer := range resp.Peers {
			fmt.Printf("%s:%d\n", peer.AddrIPV4, peer.Port)
		}

	case "handshake":

		torrentPath := os.Args[2]

		t, err := torrent.NewSingleTorrentFromFile(torrentPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		parts := strings.Split(os.Args[3], ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		peer := &torrent.Peer{
			AddrIPV4: parts[0],
			Port:     uint16(port),
		}

		pc, err := conn.EstablishConnection(torrent.LocalPeerID, peer, t, logger)
		defer pc.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Peer ID: %x\n", pc.RemotePeerID())

	case "download_piece":
		dowCmd := flag.NewFlagSet("download_piece", flag.ExitOnError)
		savePath := dowCmd.String("o", "", "Sets the output path for the piece")

		dowCmd.Parse(os.Args[2:])
		if len(dowCmd.Args()) != 2 {
			fmt.Println("Missing arguments")
			os.Exit(1)
		}

		torrentFilePath := dowCmd.Arg(0)
		pieceIndex, err := strconv.Atoi(dowCmd.Arg(1))
		if err != nil {
			fmt.Println("pieceIndex must be an integer")
			os.Exit(1)
		}

		dowService := services.NewDownloadPieceService()
		if err := dowService.DownloadPiece(*savePath, torrentFilePath, pieceIndex); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logger.Printf("Piece %d downloaded to %s.\n", pieceIndex, *savePath)

	case "download":
		fileCmd := flag.NewFlagSet("download", flag.ExitOnError)
		savePath := fileCmd.String("o", "", "Sets the output path for the downloaded file")

		fileCmd.Parse(os.Args[2:])
		if len(fileCmd.Args()) != 1 {
			fmt.Println("Invalid number of arguments provided")
			os.Exit(1)
		}
		torrentFilePath := fileCmd.Arg(0)

		downloadService := services.NewDownloadFileService()

		if err := downloadService.DownloadFile(torrentFilePath, *savePath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logger.Printf("Downloaded %s to %s\n", torrentFilePath, *savePath)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
