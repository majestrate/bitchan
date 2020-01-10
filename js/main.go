package main

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	log.WithFields(logrus.Fields{
		"state": "starting",
	}).Info("bitchan main start")
}
