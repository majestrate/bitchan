package gossip

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"github.com/majestrate/bitchan/model"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/bencode"
	"io"
	"lukechampine.com/blake3"
	"net/http"
	"net/url"
)

var log = logrus.New()

type HttpFeed struct {
	u          *url.URL
	pk         ed25519.PublicKey
	shouldQuit bool
}

func (f *HttpFeed) Stop() {
	f.shouldQuit = true
}

func newHttpFeed(u *url.URL, pk ed25519.PublicKey) *HttpFeed {
	return &HttpFeed{
		u:  u,
		pk: pk[:],
	}
}

const HttpFeedMimeType = "application/x-bitchan-metadata"

func newDecoder(r io.Reader) *bencode.Decoder {
	dec := bencode.NewDecoder(r)
	dec.SetFailOnUnorderedKeys(true)
	return dec
}

func (f *HttpFeed) verifySig(digest, sig []byte) bool {
	return ed25519.Verify(f.pk, digest, sig)
}

func encodeKey(pk ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(pk[:])
}

func doPost(remoteURL string, obj interface{}) error {
	buf := new(bytes.Buffer)
	bencode.NewEncoder(buf).Encode(obj)
	resp, err := http.Post(remoteURL, HttpFeedMimeType, buf)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":   remoteURL,
			"error": err,
		}).Error("failed to http post")
		return err
	}
	var r model.PostResponse
	defer resp.Body.Close()
	dec := newDecoder(resp.Body)
	err = dec.Decode(&r)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":   remoteURL,
			"error": err,
		}).Error("failed to read http post response")
		return err
	}
	return nil
}

func (f *HttpFeed) fetchVerified(remoteURL string, decode func(io.Reader) (interface{}, error)) (interface{}, error) {
	resp, err := http.Get(remoteURL)
	if err != nil {
		return nil, err
	}
	val := resp.Header.Get("X-Bitchan-Ed25519-B3-Signature")
	sig, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":    remoteURL,
			"error":  err,
			"header": val,
		}).Error("failed to decode signature header")
		return nil, err
	}
	defer resp.Body.Close()
	h := blake3.New(32, nil)
	r := io.TeeReader(resp.Body, h)
	val_i, err := decode(r)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":   remoteURL,
			"error": err,
		}).Error("decode failed")
		return nil, err
	}
	digest := h.Sum(nil)
	if f.verifySig(digest, sig) {
		return val_i, err
	}
	log.WithFields(logrus.Fields{
		"url": remoteURL,
		"sig": val,
		"pk":  encodeKey(f.pk),
	}).Error("signature verify failed")
	return nil, err
}

func (f *HttpFeed) FetchNeighboors() *model.PeerList {
	val, err := f.fetchVerified(f.u.String(), func(r io.Reader) (interface{}, error) {
		list := new(model.PeerList)
		dec := newDecoder(r)
		err := dec.Decode(list)
		return list, err
	})
	if err != nil {
		return nil
	}
	if val == nil {
		return nil
	}
	return val.(*model.PeerList)
}

func (f *HttpFeed) Publish(p *model.Post) {
	if f.shouldQuit {
		return
	}
	doPost(f.u.String(), p)
}
