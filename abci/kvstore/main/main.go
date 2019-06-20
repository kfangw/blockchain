package main

import (
	"fmt"
	"github.com/kfangw/blockchain/abci/kvstore"
	"github.com/kfangw/blockchain/abci/server"
	"github.com/kfangw/blockchain/abci/types"
	"github.com/kfangw/blockchain/libs/log"
	"os"
	"os/signal"
	"syscall"
)

type logger interface {
	Info(msg string, keyvals ...interface{})
}

func TrapSignal(logger logger, cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logger.Info(fmt.Sprintf("captured %v, exiting...", sig))
			if cb != nil {
				cb()
			}
			os.Exit(0)
		}
	}()
}

func runService() error {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// Create the application - in memory or persisted to disk
	var app types.Application
	app = kvstore.NewKVStoreApplication()

	// Start the listener
	srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
	if err != nil {
		return err
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		return err
	}

	// Stop upon receiving SIGTERM or CTRL-C.
	TrapSignal(logger, func() {
		// Cleanup
		srv.Stop()
	})

	// Run forever.
	select {}
}

func main() {
	runService()
}
