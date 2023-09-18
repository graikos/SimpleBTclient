package torrent

import (
	"crypto/sha1"
	"errors"
	"io"

	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bencode"
)

type SingleTorrentFile struct {
	Announce string
	Info     map[string]interface{}
}

var ErrInvalidTorrentFormat = errors.New("invalid torrent file format")
var ErrMissingInfoKeys = errors.New("missing keys from info dictionary")
var ErrInvalidValueType = errors.New("invalid value type in dictionary")

func NewSingleTorrentFile(r io.Reader) (*SingleTorrentFile, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	decodedFile, err := bencode.DecodeBencode(string(buf))
	if err != nil {
		return nil, err
	}

	fileDict, ok := decodedFile.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidTorrentFormat
	}

	torrent := &SingleTorrentFile{}
	// checking if dictionary has 'announce' key
	if _, ok := fileDict["announce"]; !ok {
		return nil, ErrInvalidTorrentFormat
	}
	if torrent.Announce, ok = fileDict["announce"].(string); !ok {
		return nil, ErrInvalidTorrentFormat
	}

	// checking if dictionary has 'info' key
	if _, ok := fileDict["info"]; !ok {
		return nil, ErrInvalidTorrentFormat
	}
	if torrent.Info, ok = fileDict["info"].(map[string]interface{}); !ok {
		return nil, ErrInvalidTorrentFormat
	}

	requiredInfoKeys := []string{"length", "name", "piece length", "pieces"}

	for _, key := range requiredInfoKeys {
		if _, ok := torrent.Info[key]; !ok {
			return nil, ErrMissingInfoKeys
		}
	}

	return torrent, nil
}

func (t *SingleTorrentFile) Length() (int, error) {
	if l, ok := t.Info["length"].(int); !ok {
		return 0, ErrInvalidValueType
	} else {
		return l, nil
	}
}

func (t *SingleTorrentFile) Name() (string, error) {
	if name, ok := t.Info["name"].(string); !ok {
		return "", ErrInvalidValueType
	} else {
		return name, nil
	}
}

func (t *SingleTorrentFile) PieceLength() (int, error) {
	if l, ok := t.Info["piece length"].(int); !ok {
		return 0, ErrInvalidValueType
	} else {
		return l, nil
	}
}

func (t *SingleTorrentFile) Pieces() ([]byte, error) {
	if p, ok := t.Info["pieces"].([]byte); !ok {
		return nil, ErrInvalidValueType
	} else {
		return p, nil
	}
}

func (t *SingleTorrentFile) InfoHash() ([]byte, error) {
	encodedInfo, err := bencode.EncodeBencodeToString(t.Info)
	if err != nil {
		return nil, err
	}
	res := sha1.Sum([]byte(encodedInfo))
	return res[:], nil
}
