package main

import (
	// Uncomment this line to pass the first stage

	"bufio"
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

		f, err := os.Open(torrentPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		torrent, err := torrent.NewSingleTorrentFile(bufio.NewReader(f))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Tracker URL:", torrent.Announce)
		l, err := torrent.Length()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Length:", l)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
