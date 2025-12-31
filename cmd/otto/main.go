package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rustyeddy/otto"
	"github.com/rustyeddy/otto/utils"
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// TODO: add command line flags
	// configure logging.
	o := otto.OttO{
		LogConfig: utils.DefaultLogConfig(),
	}
	o.Init()
	o.Start()

	// Block until we receive a signal or done channel closes
	select {
	case sig := <-sigChan:
		slog.Info("Received shutdown signal", "signal", sig)
		o.Stop()
	case <-o.Done():
		slog.Info("Server done channel closed")
		o.Shutdown()
	}

	slog.Info("OttO server stopped")
}
