package model

import (
	"crypto/ed25519"
	"github.com/zeebo/bencode"
	"lukechampine.com/blake3"
)

type Post struct {
	MetaInfoURL string `bencode:"bitchan-metainfo-url"`
	Version     string `bencode:"bitchan-version",omit-empty`
	PostedAt    int64  `bencode:"bitchan-posted-at"`
	PubKey      string `bencode:"bitchan-poster-pubkey"`
	Signature   string `bencode:"z",omit-empty`
}

func (p *Post) hashme() []byte {
	h := blake3.New(32, nil)
	enc := bencode.NewEncoder(h)
	enc.Encode(p)
	return h.Sum(nil)
}

func (p *Post) Verify() bool {
	sig := []byte(p.Signature)
	p.Signature = ""
	digest := p.hashme()
	k := ed25519.PublicKey([]byte(p.PubKey))
	if len(k) == 0 {
		return false
	}
	return ed25519.Verify(k, digest, sig)
}

func (p *Post) Sign(sk ed25519.PrivateKey) {
	p.PubKey = string(sk.Public().(ed25519.PublicKey)[:])
	p.Signature = ""
	digest := p.hashme()
	p.Signature = string(ed25519.Sign(sk, digest))
}

type PostResponse struct {
	Response string `bencode:"bitchan-post-response"`
	Version  string `bencode:"bitchan-version",omit-empty`
	Time     int64  `bencode:"bitchan-time"`
}

const DefaultPostVersion = "1.0"
