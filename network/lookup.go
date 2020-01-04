package network

import (
	"net"
	"strings"
)

func init() {
	net.DefaultResolver.PreferGo = true
}

func LookupSelf() (string, error) {
	ips, err := net.LookupIP("localhost.loki")
	if err != nil {
		return "", err
	}
	names, err := net.LookupAddr(ips[0].String())
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(names[0], "."), nil
}
