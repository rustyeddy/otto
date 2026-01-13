package logging

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// ParseLevel converts a string into a slog.Level.
func ParseLevel(s string) (slog.Level, error) {
	value := strings.ToLower(strings.TrimSpace(s))
	if value == "warning" {
		value = "warn"
	}

	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported level %q", s)
	}
}

// Build builds a slog.Logger using the provided configuration.
func Build(cfg Config) (*slog.Logger, io.Closer, *bytes.Buffer, error) {
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	level, err := ParseLevel(cfg.Level)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		writer io.Writer
		closer io.Closer
		buf    *bytes.Buffer
	)

	switch cfg.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	case "file":
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("open log file: %w", err)
		}
		writer = file
		closer = file
	case "string":
		if cfg.Buffer != nil {
			buf = cfg.Buffer
		} else {
			buf = &bytes.Buffer{}
		}
		writer = buf
	default:
		return nil, nil, nil, fmt.Errorf("unsupported output %q", cfg.Output)
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	logger := slog.New(handler)
	return logger, closer, buf, nil
}

// ApplyGlobal applies the logger and level to slog defaults.
func ApplyGlobal(logger *slog.Logger, level slog.Level) {
	if logger == nil {
		return
	}
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(level)
}
