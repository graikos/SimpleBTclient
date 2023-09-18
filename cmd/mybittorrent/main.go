package main

import (
	// Uncomment this line to pass the first stage

	"fmt"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/torrent"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {

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

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
