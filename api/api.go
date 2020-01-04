package api

import (
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/storage"
	"github.com/majestrate/bitchan/torrent"
)

type ApiServer struct {
	Torrent *torrent.Grabber
	Storage storage.Store
	Gossip  gossip.Gossiper
}

func (a *ApiServer) Stop() {
	a.Torrent.Stop()
	a.Gossip.Stop()
}

func NewAPI() *ApiServer {
	return &ApiServer{}
}
