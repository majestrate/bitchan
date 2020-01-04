package gossip

import (
	"crypto/ed25519"
	"encoding/base64"
	"github.com/majestrate/bitchan/model"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type httpGossiper struct {
	neighboors sync.Map
}

func (g *httpGossiper) forEachFeed(visit func(*HttpFeed)) {
	g.neighboors.Range(func(_, v interface{}) bool {
		visit(v.(*HttpFeed))
		return true
	})
}

func (g *httpGossiper) Stop() {
	g.forEachFeed(func(f *HttpFeed) {
		f.Stop()
	})
}

func (g *httpGossiper) ForEachPeer(visit func(model.Peer)) {
	g.forEachFeed(func(feed *HttpFeed) {
		visit(model.Peer{
			URL: feed.u.String(),
		})
	})
}

func (g *httpGossiper) Bootstrap() {
	g.forEachFeed(func(feed *HttpFeed) {
		go func() {
			l := feed.FetchNeighboors()
			if l == nil {
				return
			}
			for name, peer := range l.Peers {
				u, _ := url.Parse(peer.URL)
				if u == nil {
					continue
				}
				if u.Host == name {
					g.AddNeigboor(u)
				}
			}
		}()
	})
}

func newHttpGossiper() *httpGossiper {
	return new(httpGossiper)
}

func (g *httpGossiper) BroadcastLocalPost(p *model.Post) {
	g.forEachFeed(func(feed *HttpFeed) {
		go feed.Publish(p)
	})
}

func (g *httpGossiper) AddNeigboor(n *url.URL) {
	_, has := g.neighboors.Load(n.Host)
	if !has {
		// read pubkey
		resp, err := http.Get(n.String())
		if err != nil {
			return
		}
		defer resp.Body.Close()
		dec := base64.NewDecoder(base64.URLEncoding, resp.Body)
		var pk ed25519.PublicKey
		_, err = io.ReadFull(dec, pk[:])
		if err != nil {
			return
		}
		g.neighboors.Store(n.Host, newHttpFeed(n, pk))
	}
}
