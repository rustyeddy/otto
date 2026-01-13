package logging

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	DefaultLevel  = "info"
	DefaultFormat = "text"
	DefaultOutput = "stdout"
)

// Config defines configuration for logging outputs and formatting.
type Config struct {
	Level    string        `json:"level"`
	Format   string        `json:"format"`
	Output   string        `json:"output"`
	FilePath string        `json:"filePath,omitempty"`
	Buffer   *bytes.Buffer `json:"-"`
}

// DefaultConfig returns the default logging configuration.
func DefaultConfig() Config {
	return Config{
		Level:  DefaultLevel,
		Format: DefaultFormat,
		Output: DefaultOutput,
	}
}

// WithDefaults fills in empty fields with defaults.
func (c Config) WithDefaults() Config {
	if strings.TrimSpace(c.Level) == "" {
		c.Level = DefaultLevel
	}
	if strings.TrimSpace(c.Format) == "" {
		c.Format = DefaultFormat
	}
	if strings.TrimSpace(c.Output) == "" {
		c.Output = DefaultOutput
	}
	return c
}

// Normalize lowercases string fields and clears file/buffer fields when not used.
func (c Config) Normalize() Config {
	c.Level = strings.ToLower(strings.TrimSpace(c.Level))
	c.Format = strings.ToLower(strings.TrimSpace(c.Format))
	c.Output = strings.ToLower(strings.TrimSpace(c.Output))
	if c.Output != "file" {
		c.FilePath = ""
	}
	if c.Output != "string" {
		c.Buffer = nil
	}
	return c
}

// Validate checks the configuration for supported values.
func (c Config) Validate() error {
	if _, err := ParseLevel(c.Level); err != nil {
		return err
	}

	switch c.Format {
	case "text", "json":
	default:
		return fmt.Errorf("unsupported format %q", c.Format)
	}

	switch c.Output {
	case "stdout", "stderr", "file", "string":
	default:
		return fmt.Errorf("unsupported output %q", c.Output)
	}

	if c.Output == "file" && strings.TrimSpace(c.FilePath) == "" {
		return fmt.Errorf("file output requires filePath")
	}
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	cfg = cfg.WithDefaults().Normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
