package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rustyeddy/otto"
	"github.com/rustyeddy/otto/station"
	"github.com/rustyeddy/otto/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		stationsData := []*station.StationSummary{
			{
				ID:        "station-01",
				Hostname:  "station-01-hostname",
				LastHeard: 2 * time.Minute,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stationsData)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	stations, err := client.GetStations()
	require.NoError(t, err)
	assert.Equal(t, 1, len(stations))
	st := stations[0]
	assert.Equal(t, "station-01", st.ID)
	assert.Equal(t, "station-01-hostname", st.Hostname)
	assert.Equal(t, 2*time.Minute, st.LastHeard)
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
	assert.NotNil(t, deflc.FilePath, lc.FilePath)
	assert.NotNil(t, deflc.Format, lc.Format)
	assert.NotNil(t, deflc.Level, lc.Level)
	assert.NotNil(t, deflc.Output, lc.Output)
}
