package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rustyeddy/otto"
	"github.com/rustyeddy/otto/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8011")
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.BaseURL != "http://localhost:8011" {
		t.Errorf("Expected BaseURL to be http://localhost:8011, got %s", client.BaseURL)
	}
}

func TestGetStats(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stats" {
			t.Errorf("Expected path /api/stats, got %s", r.URL.Path)
		}
		stats := map[string]interface{}{
			"Goroutines": 10,
			"CPUs":       4,
			"GoVersion":  "go1.21.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	stats, err := client.GetStats()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats["Goroutines"] != float64(10) {
		t.Errorf("Expected Goroutines to be 10, got %v", stats["Goroutines"])
	}
}

func TestPing(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" {
			t.Errorf("Expected path /ping, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	err := client.Ping()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestGetStats_ServerError(t *testing.T) {
	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	_, err := client.GetStats()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetStations(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stations" {
			t.Errorf("Expected path /api/stations, got %s", r.URL.Path)
		}
		stationsData := map[string]interface{}{
			"stations": map[string]interface{}{
				"station1": map[string]interface{}{
					"id":       "station1",
					"hostname": "test-host",
				},
			},
			"stale": map[string]interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stationsData)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	stations, err := client.GetStations()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	stationsMap, ok := stations["stations"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected stations map in response")
	}

	if len(stationsMap) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stationsMap))
	}
}

func TestGetStations_ServerError(t *testing.T) {
	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	_, err := client.GetStations()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestGetVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/version")
		version := map[string]any{
			"version": otto.Version,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(version)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	vmap, err := client.GetVersion()
	assert.NoError(t, err)
	assert.Equal(t, vmap["version"], otto.Version)
}

func TestGetLogConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.String(), "/api/log")

		lc := utils.DefaultLogConfig()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lc)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	lc, err := client.GetLogConfig()
	assert.NoError(t, err)
	deflc := utils.DefaultLogConfig()
	assert.NotNil(t, deflc.FilePath, lc["FilePath"])
	assert.NotNil(t, deflc.Format, "text")
	assert.NotNil(t, deflc.Level, "info")
	assert.NotNil(t, deflc.Output, "stdout")
}
