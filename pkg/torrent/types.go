package torrent

type TrackerResponse struct {
	Interval int
	Peers    []*Peer
}
