package web

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/majestrate/bitchan/api"
	"github.com/majestrate/bitchan/gossip"
	"github.com/majestrate/bitchan/model"
	"github.com/zeebo/bencode"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MiddleWare struct {
	Api      *api.ApiServer
	router   *gin.Engine
	privkey  ed25519.PrivateKey
	self     model.Peer
	hostname string
	port     string
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

func newDecoder(r io.Reader) *bencode.Decoder {
	dec := bencode.NewDecoder(r)
	dec.SetFailOnUnorderedKeys(true)
	return dec
}

func (m *MiddleWare) makeFilesURL(fname string) string {
	return "http://" + net.JoinHostPort(m.hostname, m.port) + "/files/" + filepath.Base(fname)
}

func mktmp(ext string) string {
	now := time.Now().UnixNano()
	var b [4]byte
	rand.Read(b[:])
	r := strings.Trim(base64.URLEncoding.EncodeToString(b[:]), "=")
	return filepath.Join(os.TempDir(), fmt.Sprintf("%d-%s%s", now, r, ext))
}

func (m *MiddleWare) makeAdminPost(hdr *multipart.FileHeader) (p *model.Post, err error) {
	h := sha256.New()
	ext := filepath.Ext(hdr.Filename)
	tmpfile := mktmp(ext)
	inf, err := hdr.Open()
	if err != nil {
		return nil, err
	}
	defer inf.Close()
	of, err := os.Create(tmpfile)
	if err != nil {
		return nil, err
	}
	r := io.TeeReader(inf, h)
	_, err = io.Copy(of, r)
	of.Close()
	if err != nil {
		os.Remove(tmpfile)
		return nil, err
	}
	d := h.Sum(nil)
	fname := base64.URLEncoding.EncodeToString(d[:]) + ext
	fname = filepath.Join(m.Api.Storage.GetRoot(), fname)
	err = os.Rename(tmpfile, fname)
	if err != nil {
		os.Remove(tmpfile)
		return nil, err
	}
	torrentFile := fname + ".torrent"
	err = m.Api.MakeTorrent(fname, torrentFile)
	if err != nil {
		os.Remove(fname)
		os.Remove(torrentFile)
		return nil, err
	}
	now := time.Now().UnixNano()
	p = &model.Post{
		MetaInfoURL: m.makeFilesURL(torrentFile),
		PostedAt:    now,
	}
	p.Sign(m.privkey)
	return p, nil
}

func (m *MiddleWare) SetupRoutes() {
	// sendresult sends signed result
	sendResult := func(c *gin.Context, buf *bytes.Buffer, ct string) {
		sig := ed25519.Sign(m.privkey, buf.Bytes())
		c.Header("X-Bitchan-Ed25519-Signature", encodeSig(sig))
		c.Header("Content-Type", ct)
		c.String(http.StatusOK, buf.String())
	}

	m.router.StaticFS("/files", http.Dir(m.Api.Storage.GetRoot()))

	m.router.GET("/bitchan/v1/admin/add-peer", func(c *gin.Context) {
		rhost, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		rip := net.ParseIP(rhost)
		if !rip.IsLoopback() {
			// deny
			c.String(http.StatusForbidden, "nah")
			return
		}
		u := c.DefaultQuery("url", "")
		if u == "" {
			c.String(http.StatusBadRequest, "no url provided")
			return
		}
		remote, err := url.Parse(u)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		if m.Api.Gossip.AddNeighboor(remote) {
			c.String(http.StatusCreated, "added")
		} else {
			c.String(http.StatusBadRequest, "not added")
		}
	})

	m.router.POST("/bitchan/v1/admin/post", func(c *gin.Context) {
		rhost, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
		rip := net.ParseIP(rhost)
		if !rip.IsLoopback() {
			// deny
			c.String(http.StatusForbidden, "nah")
			return
		}
		f, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		p, err := m.makeAdminPost(f)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		m.Api.Gossip.BroadcastLocalPost(p)
		c.String(http.StatusCreated, "posted")
	})

	m.router.GET("/bitchan/v1/admin/bootstrap", func(c *gin.Context) {
		rhost, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
		rip := net.ParseIP(rhost)
		if !rip.IsLoopback() {
			// deny
			c.String(http.StatusForbidden, "nah")
			return
		}
		go m.Api.Gossip.Bootstrap()
		c.String(http.StatusCreated, "bootstrap started")
	})

	m.router.GET("/bitchan/v1/self", func(c *gin.Context) {
		buf := new(bytes.Buffer)
		enc := bencode.NewEncoder(buf)
		enc.Encode(m.self)
		sendResult(c, buf, gossip.HttpFeedMimeType)
	})
	m.router.GET("/bitchan/v1/pubkey", func(c *gin.Context) {
		pk := m.privkey.Public().(ed25519.PublicKey)
		c.Header("Content-Type", BitchanPubKeyContentType)
		c.String(http.StatusOK, encodePubKey(pk))
	})
	m.router.GET("/bitchan/v1/peer", func(c *gin.Context) {
		port := c.DefaultQuery("port", "8800")
		rhost, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		host := c.DefaultQuery("host", "")
		if host == "" {
			names, err := net.LookupAddr(rhost)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			host = strings.TrimSuffix(names[0], ".")
		} else {
			addrs, err := net.LookupIP(host)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			found := false
			for _, addr := range addrs {
				if addr.String() == rhost {
					found = true
				}
			}
			if !found {
				c.String(http.StatusForbidden, "spoofed name")
				return
			}
		}
		fedurl := "http://" + net.JoinHostPort(host, port) + "/bitchan/v1/federate"
		u, err := url.Parse(fedurl)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		if m.Api.Gossip.AddNeighboor(u) {
			c.String(http.StatusCreated, "")
		} else {
			c.String(http.StatusForbidden, "not added")
		}
	})
	m.router.POST("/bitchan/v1/federate", func(c *gin.Context) {
		ct := c.Request.Header.Get("Content-Type")
		if ct != gossip.HttpFeedMimeType {
			c.String(http.StatusForbidden, "")
			return
		}

		var p model.Post
		defer c.Request.Body.Close()
		dec := newDecoder(c.Request.Body)
		err := dec.Decode(&p)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		if !p.Verify() {
			c.String(http.StatusForbidden, "bad post signature")
			return
		}
		err = m.Api.Torrent.Grab(p.MetaInfoURL)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		var result model.PostResponse
		result.Response = "accepted"
		result.Time = time.Now().UnixNano()
		buf := new(bytes.Buffer)
		enc := bencode.NewEncoder(buf)
		enc.Encode(result)
		sendResult(c, buf, gossip.HttpFeedMimeType)
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

		list.Peers[m.hostname] = m.self
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

func New(host string, port string) *MiddleWare {
	m := &MiddleWare{
		Api:      nil,
		router:   gin.Default(),
		hostname: host,
		port:     port,
		self: model.Peer{
			URL: "http://" + net.JoinHostPort(host, port) + "/bitchan/v1/federate",
		},
	}
	return m
}
