package torrent

import (
	"github.com/anacrolix/torrent"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/storage"
)

type Grabber struct {
	client   *torrent.Client
	gossiper gossip.Gossiper
	store    storage.Store
}

func (g *Grabber) Stop() {
	g.client.Close()
}

func NewGrabber(st storage.Store, g gossip.Gossiper) *Grabber {
	t, _ := torrent.NewClient(torrent.NewDefaultClientConfig())
	return &Grabber{
		client:   t,
		gossiper: g,
		store:    st,
	}
}
