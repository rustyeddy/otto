package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rustyeddy/otto"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start oTTo the Server",
	Long:  `Start OttO the IoT Server`,
	Run:   serveRun,
}

var (
	foreground bool
)

func init() {
	serveCmd.Flags().BoolVar(&foreground, "foreground", false, "Run the server command in the foreground")
}

func serveRun(cmd *cobra.Command, args []string) {
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
