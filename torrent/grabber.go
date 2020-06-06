package torrent

import (
	"errors"
	alog "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/storage"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var log = logrus.New()

var ErrNotAdded = errors.New("torrent not addded")

type Grabber struct {
	Client   *torrent.Client
	gossiper gossip.Gossiper
	store    storage.Store
}

func (g *Grabber) ForEachSeed(visit func(*torrent.Torrent)) {
	for _, t := range g.Client.Torrents() {
		if t.Seeding() {
			visit(t)
		}
	}
}

func (g *Grabber) Grab(metainfoURL string) error {
	log.WithFields(logrus.Fields{
		"url": metainfoURL,
	}).Info("grabbing metainfo")
	resp, err := http.Get(metainfoURL)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url": metainfoURL,
			"err": err,
		}).Error("grabbing metainfo failed")
		return err
	}
	defer resp.Body.Close()
	mi, err := metainfo.Load(resp.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url": metainfoURL,
			"err": err,
		}).Error("reading metainfo failed")
		return err
	}
	t, err := g.Client.AddTorrent(mi)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url": metainfoURL,
			"err": err,
		}).Error("failed to add torrent")
		return err
	}

	log.WithFields(logrus.Fields{
		"url": metainfoURL,
	}).Info("download starting")
	u, _ := url.Parse(metainfoURL)
	f, err := os.Create(filepath.Join(g.store.GetRoot(), filepath.Base(u.Path)))
	if err == nil {
		defer f.Close()
		mi.Write(f)
	}
	if t.Seeding() {
		log.WithFields(logrus.Fields{
			"url": metainfoURL,
			"infohash": t.InfoHash().HexString(),
		}).Info("seeding")
	}else {
		log.WithFields(logrus.Fields{
			"url": metainfoURL,
			"infohash": t.InfoHash().HexString(),
		}).Info("downloading")
		t.DownloadAll()
	}
	return nil
}

func (g *Grabber) Stop() {
	g.Client.Close()
}

func NewGrabber(st storage.Store, g gossip.Gossiper) *Grabber {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = st.GetRoot()
	cfg.Seed = true
	cfg.Debug = false
	cfg.Logger = alog.Discard
	t, _ := torrent.NewClient(cfg)
	return &Grabber{
		Client:   t,
		gossiper: g,
		store:    st,
	}
}
