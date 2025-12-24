package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rustyeddy/otto/utils"
)

// StatsHandler handles REST API requests for runtime statistics
type StatsHandler struct{}

// ServeHTTP implements http.Handler to return runtime statistics as JSON
func (h StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats := utils.GetStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("Failed to encode stats", "error", err)
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
		return
	}
}
