package utils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	logfile string = "otto.log"
)

// LogOutput defines where logs should be written
type LogOutput string

const (
	LogOutputFile   LogOutput = "file"
	LogOutputStdout LogOutput = "stdout"
	LogOutputStderr LogOutput = "stderr"
	LogOutputString LogOutput = "string"
)

// LogFormat defines the format of log output
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// LogConfig holds the configuration for logging
type LogConfig struct {
	Level      string    // Log level: debug, info, warn, error
	Output     LogOutput // Output destination: file, stdout, stderr, string
	Format     LogFormat // Format: text, json
	FilePath   string    // Path to log file (used when Output is file)
	Buffer     *bytes.Buffer // Buffer to write logs to (used when Output is string)
}

// InitLogger initializes the logger with the old signature for backward compatibility
func InitLogger(lstr string, lf string) {
	if lf == "" {
		lf = logfile
	}
	level := SetLogLevel(lstr)
	f, err := os.OpenFile(lf, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("error opening log ", "err", err)
	}
	l := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(l)
}

// InitLoggerWithConfig initializes the logger with a LogConfig
func InitLoggerWithConfig(config LogConfig) (*bytes.Buffer, error) {
	level := SetLogLevel(config.Level)
	
	var writer io.Writer
	var buffer *bytes.Buffer
	var err error
	
	// Determine output writer
	switch config.Output {
	case LogOutputFile:
		filePath := config.FilePath
		if filePath == "" {
			filePath = logfile
		}
		writer, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("error opening log file: %w", err)
		}
	case LogOutputStdout:
		writer = os.Stdout
	case LogOutputStderr:
		writer = os.Stderr
	case LogOutputString:
		if config.Buffer != nil {
			buffer = config.Buffer
		} else {
			buffer = &bytes.Buffer{}
		}
		writer = buffer
	default:
		writer = os.Stdout
	}
	
	// Create handler based on format
	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{Level: level}
	
	switch config.Format {
	case LogFormatJSON:
		handler = slog.NewJSONHandler(writer, handlerOpts)
	case LogFormatText:
		handler = slog.NewTextHandler(writer, handlerOpts)
	default:
		handler = slog.NewTextHandler(writer, handlerOpts)
	}
	
	logger := slog.New(handler)
	slog.SetDefault(logger)
	
	return buffer, nil
}

func SetLogLevel(loglevel string) slog.Level {
	var level slog.Level

	switch loglevel {
	case "debug":
		level = slog.LevelDebug

	case "info":
		level = slog.LevelInfo

	case "warn":
		level = slog.LevelWarn

	case "error":
		level = slog.LevelError

	default:
		fmt.Printf("unknown loglevel %s sticking with warn", loglevel)
	}
	slog.SetLogLoggerLevel(level)
	return level
}
