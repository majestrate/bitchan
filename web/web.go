package web

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/majestrate/bitchan/api"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/model"
	"github.com/zeebo/bencode"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type MiddleWare struct {
	Api     *api.ApiServer
	router  *gin.Engine
	privkey ed25519.PrivateKey
}

func (m *MiddleWare) EnsureKeyFile(fname string) error {
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		err = nil
		seed := make([]byte, 32)
		n, err := rand.Read(seed)
		if n != 32 || err != nil {
			return err
		}
		err = ioutil.WriteFile(fname, seed, os.FileMode(0600))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	m.privkey = ed25519.NewKeyFromSeed(data)
	return nil
}

func (m *MiddleWare) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.router.ServeHTTP(w, r)
}

const BitchanPubKeyContentType = "application/x-bitchan-identity"

var b64enc = base64.StdEncoding

func encodeSig(sig []byte) string {
	return b64enc.EncodeToString(sig[:])
}

func encodePubKey(k ed25519.PublicKey) string {
	return b64enc.EncodeToString(k[:])
}

func (m *MiddleWare) SetupRoutes() {
	// sendresult sends signed result
	sendResult := func(c *gin.Context, buf *bytes.Buffer, ct string) {
		sig := ed25519.Sign(m.privkey, buf.Bytes())
		c.Header("X-Bitchan-Ed25519-Signature", encodeSig(sig))
		c.Header("Content-Type", ct)
		c.String(http.StatusOK, buf.String())
	}

	m.router.GET("/bitchan/v1/pubkey", func(c *gin.Context) {
		pk := m.privkey.Public().(ed25519.PublicKey)
		c.Header("Content-Type", BitchanPubKeyContentType)
		c.String(http.StatusOK, encodePubKey(pk))
	})
	m.router.GET("/bitchan/v1/federate", func(c *gin.Context) {
		var list model.PeerList
		list.Peers = make(map[string]model.Peer)
		list.Time = time.Now().UnixNano()
		m.Api.Gossip.ForEachPeer(func(p model.Peer) {
			u, _ := url.Parse(p.URL)
			if u == nil {
				return
			}
			list.Peers[u.Host] = p
		})

		buf := new(bytes.Buffer)
		enc := bencode.NewEncoder(buf)
		err := enc.Encode(list)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		sendResult(c, buf, gossip.HttpFeedMimeType)
	})
}

func (m *MiddleWare) Stop() {
	m.Api.Stop()
}

func New() *MiddleWare {
	m := &MiddleWare{
		Api:    nil,
		router: gin.Default(),
	}
	m.SetupRoutes()
	return m
}
