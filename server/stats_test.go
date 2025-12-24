package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatsHandler(t *testing.T) {
	handler := StatsHandler{}

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that we have expected fields
	if _, ok := stats["Goroutines"]; !ok {
		t.Error("Expected Goroutines field in stats")
	}
	if _, ok := stats["CPUs"]; !ok {
		t.Error("Expected CPUs field in stats")
	}
	if _, ok := stats["GoVersion"]; !ok {
		t.Error("Expected GoVersion field in stats")
	}
}
