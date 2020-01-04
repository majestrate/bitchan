package torrent

import (
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/storage"
	"github.com/sirupsen/logrus"
	"net/http"
)

var log = logrus.New()

var ErrNotAdded = errors.New("torrent not addded")

type Grabber struct {
	Client   *torrent.Client
	gossiper gossip.Gossiper
	store    storage.Store
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
	t.DownloadAll()
	return nil
}

func (g *Grabber) Stop() {
	g.Client.Close()
}

func NewGrabber(st storage.Store, g gossip.Gossiper) *Grabber {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = st.GetRoot()
	cfg.Seed = true
	t, _ := torrent.NewClient(cfg)
	return &Grabber{
		Client:   t,
		gossiper: g,
		store:    st,
	}
}
