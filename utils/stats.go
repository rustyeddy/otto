package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"
)

type Stats struct {
	Goroutines int
	CPUs       int
	runtime.MemStats
	GoVersion string
}

func GetStats() *Stats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s := &Stats{
		Goroutines: runtime.NumGoroutine(),
		CPUs:       runtime.NumCPU(),
		MemStats:   m,
		GoVersion:  runtime.Version(),
	}

	return s
}

// ServeHTTP implements http.Handler to return runtime statistics as JSON
func (h *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats := h
	if stats == nil {
		stats = GetStats()
	}

	// Encode to bytes first to avoid setting Content-Type header before knowing if encoding succeeds.
	// If encoding fails and we've already set Content-Type to "application/json", the error response
	// would have the wrong content type (error responses from http.Error are plain text).
	data, err := json.Marshal(stats)
	if err != nil {
		slog.Error("Failed to encode stats", "error", err)
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
		return
	}

	// Only set Content-Type after successful encoding
	w.Header().Set("Content-Type", "application/json")
	// Note: Once headers are written, we cannot send an error response if Write fails
	if _, err := w.Write(data); err != nil {
		slog.Error("Failed to write stats response", "error", err)
	}
}
