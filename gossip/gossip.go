package gossip

import (
	"github.com/majestrate/bitchan/model"
	"net/url"
)

type Gossiper interface {
	BroadcastLocalPost(*model.Post)
	AddNeighboor(u *url.URL) bool
	Stop()
	Bootstrap()
	ForEachPeer(func(model.Peer))
}

func NewServer(hostname string) Gossiper {
	return newHttpGossiper(hostname)
}
