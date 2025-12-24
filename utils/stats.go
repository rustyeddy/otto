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

// ServeHTTP implements http.Handler to return runtime statistics as JSON.
// It always calls GetStats() to get fresh runtime statistics.
func (h *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats := GetStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode stats", "error", err)
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
		return
	}
}
