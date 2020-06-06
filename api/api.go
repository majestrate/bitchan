package api

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/storage"
	"github.com/majestrate/bitchan/torrent"
	"os"
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

func (a *ApiServer) MakeTorrent(rootf, outf string) (string, error) {
	var mi metainfo.MetaInfo
	mi.Announce = "udp://opentracker.i2p.rocks:6969/announce"
	var miInfo metainfo.Info
	miInfo.PieceLength = 128 * 1024
	err := miInfo.BuildFromFilePath(rootf)
	if err != nil {
		return "", err
	}
	mi.InfoBytes, err = bencode.Marshal(miInfo)
	if err != nil {
		return "", err
	}
	f, err := os.Create(outf)
	if err != nil {
		return "", err
	}
	infohash_hex := mi.HashInfoBytes().HexString()
	defer f.Close()
	return infohash_hex, mi.Write(f)
}

func NewAPI() *ApiServer {
	return &ApiServer{}
}
