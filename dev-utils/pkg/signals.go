package devutil

import (
	"os"
	"os/signal"
	"syscall"
)

func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		// first signal, close
		close(stop)
		<-c
		// second signal, hard exit
		os.Exit(1)
	}()

	return stop
}
