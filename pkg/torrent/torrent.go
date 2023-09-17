package torrent

type SingleTorrentFile struct {
	Announce string
	Info     map[string]interface{}
}
