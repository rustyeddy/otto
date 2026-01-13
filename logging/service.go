package logging

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// Service manages the logging configuration and exposes an HTTP API.
type Service struct {
	mu     sync.Mutex
	cfg    Config
	closer io.Closer
}

// NewService creates a new Service and applies the configuration.
func NewService(cfg Config) (*Service, error) {
	svc := &Service{}
	if err := svc.SetConfig(cfg); err != nil {
		return nil, err
	}
	return svc, nil
}

// Config returns the current configuration.
func (s *Service) Config() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

// SetConfig applies a new configuration and updates the global logger.
func (s *Service) SetConfig(cfg Config) error {
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}

	level, err := ParseLevel(cfg.Level)
	if err != nil {
		return err
	}

	logger, closer, buf, err := Build(cfg)
	if err != nil {
		return err
	}

	ApplyGlobal(logger, level)

	s.mu.Lock()
	oldCloser := s.closer
	cfg.Buffer = buf
	if cfg.Output != "string" {
		cfg.Buffer = nil
	}
	s.cfg = cfg
	s.closer = closer
	s.mu.Unlock()

	if oldCloser != nil {
		_ = oldCloser.Close()
	}

	return nil
}

// ServeHTTP serves the logging configuration endpoint.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.Config()
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPut:
		var cfg Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := s.SetConfig(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, s.Config())
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{Error: err.Error()})
}
