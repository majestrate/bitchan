package gossip

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"github.com/majestrate/bitchan/model"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/bencode"
	"io"
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
		pk: pk,
	}
}

const HttpFeedMimeType = "application/x-bitchan-metadata"

func newDecoder(r io.Reader) *bencode.Decoder {
	dec := bencode.NewDecoder(r)
	dec.SetFailOnUnorderedKeys(true)
	return dec
}

func (f *HttpFeed) verifySig(body *bytes.Buffer, sig []byte) bool {
	return ed25519.Verify(f.pk, body.Bytes(), sig)
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

func (f *HttpFeed) FetchNeighboors() *model.PeerList {
	var list model.PeerList
	remoteURL := f.u.String()
	resp, err := http.Get(remoteURL)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":   remoteURL,
			"error": err,
		}).Error("failed to http get")
		return nil
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != HttpFeedMimeType {

		log.WithFields(logrus.Fields{
			"url":          remoteURL,
			"content-type": contentType,
		}).Error("bad content type")
		return nil
	}
	val := resp.Header.Get("X-Bitchan-Ed25519-Signature")
	sig, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":    remoteURL,
			"error":  err,
			"header": val,
		}).Error("failed to decode signature header")
		return nil
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)
	dec := newDecoder(buf)
	err = dec.Decode(&list)
	if err != nil {
		log.WithFields(logrus.Fields{
			"url":   remoteURL,
			"error": err,
		}).Error("decode failed")
		return nil
	}
	if f.verifySig(buf, sig) {
		return &list
	}
	log.WithFields(logrus.Fields{
		"url": remoteURL,
		"sig": val,
		"pk":  f.pk,
	}).Error("signature verify failed")
	return nil

}

func (f *HttpFeed) Publish(p *model.Post) {
	if f.shouldQuit {
		return
	}
	doPost(f.u.String(), p)
}
