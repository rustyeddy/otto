package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rustyeddy/otto/logging"
	"github.com/spf13/cobra"
)

const serverAddr = ":8011"

var (
	logLevel  string
	logFormat string
	logOutput string
	logFile   string
)

var rootCmd = &cobra.Command{
	Use:           "otto",
	Short:         "OttO IoT server",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&logLevel, "log-level", logging.DefaultLevel, "Log level (debug, info, warn, error)")
	serveCmd.Flags().StringVar(&logFormat, "log-format", logging.DefaultFormat, "Log format (text, json)")
	serveCmd.Flags().StringVar(&logOutput, "log-output", logging.DefaultOutput, "Log output (stdout, stderr, file, string)")
	serveCmd.Flags().StringVar(&logFile, "log-file", "", "Log file path (required when log-output=file)")
	rootCmd.AddCommand(serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	if strings.EqualFold(logOutput, "file") && strings.TrimSpace(logFile) == "" {
		return errors.New("log-output=file requires --log-file")
	}

	cfg := logging.Config{
		Level:    logLevel,
		Format:   logFormat,
		Output:   logOutput,
		FilePath: logFile,
	}

	logService, err := logging.NewService(cfg)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/log", logService)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
