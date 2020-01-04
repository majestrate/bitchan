package main

import (
	"github.com/majestrate/bitchan/api"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/network"
	"github.com/majestrate/bitchan/signals"
	"github.com/majestrate/bitchan/storage"
	"github.com/majestrate/bitchan/torrent"
	"github.com/majestrate/bitchan/web"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var log = logrus.New()

func main() {
	var err error
	var host string
	if len(os.Args) == 1 {
		host, err = network.LookupSelf()
	} else {
		host = os.Args[1]
	}
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to resolve our hostname")
	}
	log.WithFields(logrus.Fields{
		"hostname": host,
	}).Info("set hostname")

	h := web.New(host)
	h.EnsureKeyFile("identity.key")
	h.Api = api.NewAPI()

	h.Api.Storage = storage.NewStorage()
	h.Api.Storage.SetRoot("storage")
	h.Api.Gossip = gossip.NewServer()
	h.Api.Torrent = torrent.NewGrabber(h.Api.Storage, h.Api.Gossip)

	s := &http.Server{
		Addr:    "127.0.0.1:8800",
		Handler: h,
	}
	signals.SetupSignals(func() {

	}, func() {
		s.Close()
		h.Stop()
	})
	go func() {
		log.Infof("staring up...")
		s.ListenAndServe()
	}()
	signals.Wait()
}
