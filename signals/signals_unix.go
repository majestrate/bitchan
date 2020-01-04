package signals

import (
	"os"
	"os/signal"
	"syscall"
)

var chnl = make(chan os.Signal)

var onExit ExitFunc
var onReload ReloadFunc

func SetupSignals(reload ReloadFunc, exit ExitFunc) {
	onExit = exit
	onReload = reload
	signal.Notify(chnl, os.Interrupt, syscall.SIGHUP)
}

func Wait() {
	for {
		sig := <-chnl
		switch sig {
		case os.Interrupt:
			if onExit != nil {
				onExit()
				return
			}
		case syscall.SIGHUP:
			if onReload != nil {
				onReload()
			}
		}
	}
}
