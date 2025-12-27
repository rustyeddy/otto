package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rustyeddy/otto"
)

func main() {
	o := otto.OttO{}
	o.Init()
	o.Start()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or done channel closes
	select {
	case sig := <-sigChan:
		slog.Info("Received shutdown signal", "signal", sig)
		o.Stop()
	case <-o.Done():
		slog.Info("Server done channel closed")
	}

	slog.Info("OttO server stopped")
}
