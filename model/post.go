package model

type Post struct {
	MetaInfoURL string `bencode:"bitchan-metainfo-url"`
	Version     string `bencode:"bitchan-version"`
	PostedAt    int64  `bencode:"bitchan-posted-at"`
	Signature   string `bencode:"Z",omit-empty`
}

type PostResponse struct {
	Response string `bencode:"bitchan-post-response"`
	Version  string `bencode:"bitchan-version"`
	Time     int64  `bencode:"bitchan-time"`
}

const DefaultPostVersion = "1.0"
