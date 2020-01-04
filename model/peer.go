package model

// Peer contains info about a peer
type Peer struct {
	URL string `bencode:"bitchan-peer-url"`
}

// PeerList maps hostname to peer url
type PeerList struct {
	Peers map[string]Peer `bencode:"bitchan-peers"`
	Time  int64           `bencode:"bitchan-time"`
}
