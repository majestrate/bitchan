package gossip

import (
	"crypto/ed25519"
	"encoding/base64"
	"github.com/majestrate/bitchan/model"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type httpGossiper struct {
	neighboors sync.Map
	hostname   string
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
					g.AddNeighboor(u)
				}
			}
		}()
	})
}

func newHttpGossiper(hostname string) *httpGossiper {
	return &httpGossiper{
		hostname: hostname,
	}
}

func (g *httpGossiper) BroadcastLocalPost(p *model.Post) {
	g.forEachFeed(func(feed *HttpFeed) {
		go feed.Publish(p)
	})
}

func (g *httpGossiper) AddNeighboor(n *url.URL) bool {
	_, has := g.neighboors.Load(n.Host)
	if has {
		log.WithFields(logrus.Fields{
			"host": n.Host,
		}).Error("already have neighboor")
		return false
	} else {
		// read pubkey
		pkurl, _ := url.Parse(n.String())
		pkurl.Path = "/bitchan/v1/pubkey"
		resp, err := http.Get(pkurl.String())
		if err != nil {

			log.WithFields(logrus.Fields{
				"url":   pkurl.String(),
				"error": err,
			}).Error("failed to do request")
			return false
		}
		defer resp.Body.Close()
		dec := base64.NewDecoder(base64.StdEncoding, resp.Body)
		pk := make(ed25519.PublicKey, 32)
		_, err = io.ReadFull(dec, pk[:])
		if err != nil {
			log.WithFields(logrus.Fields{
				"url":   pkurl.String(),
				"error": err,
			}).Error("failed to read pubkey")
			return false
		}
		g.neighboors.Store(n.Host, newHttpFeed(n, pk))
		addme, _ := url.Parse(n.String())
		addme.Path = "/bitchan/v1/peer?host=" + g.hostname
		_, err = http.Get(addme.String())
		if err == nil {
			log.WithFields(logrus.Fields{
				"host": n.Host,
			}).Info("added neighboor")
		}
		return true
	}
}
