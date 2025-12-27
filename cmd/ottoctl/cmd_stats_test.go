package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rustyeddy/otto/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestStatsCmd(t *testing.T) {
	// Create a buffer to capture the output
	output := new(bytes.Buffer)
	statsCmd.SetOut(output)
	cmdOutput = output

	// Execute the command
	statsCmd.Run(statsCmd, []string{})

	// Check if the output contains the expected string
	got := output.String()
	if got == "" {
		t.Errorf("Expected output, but got empty string")
	}
}

func TestStatsRun(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the statsRun function
	statsRun(cmd, args)

	// No assertions here since statsRun only prints output,
	// but you can extend this test if you mock utils.GetStats().
	assert.NotEmpty(t, output.Bytes())
}

func TestStatsRun_RemoteMode(t *testing.T) {
	// Create a test server that returns stats
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

	// Save original values
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Set server URL to test server
	serverURL = ts.URL

	// Create output buffer
	output := new(bytes.Buffer)
	cmdOutput = output

	// Run the command
	cmd := &cobra.Command{}
	statsRun(cmd, []string{})

	// Check output contains JSON response
	result := output.String()
	if result == "" {
		t.Error("Expected output, got empty string")
	}
	if !bytes.Contains(output.Bytes(), []byte("Goroutines")) {
		t.Error("Expected output to contain Goroutines")
	}
}

func TestStatsRun_RemoteMode_Error(t *testing.T) {
	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	// Save original values
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Set server URL to test server
	serverURL = ts.URL

	// Create output buffer
	output := new(bytes.Buffer)
	cmdOutput = output

	// Run the command
	cmd := &cobra.Command{}
	statsRun(cmd, []string{})

	// Check output contains error message
	if !bytes.Contains(output.Bytes(), []byte("Error")) {
		t.Error("Expected output to contain error message")
	}
}

func TestStatsRun_RemoteMode_EnvVar(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"Goroutines": 5,
			"CPUs":       2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))
	defer ts.Close()

	// Save original values
	oldServerURL := serverURL
	oldEnvVar := os.Getenv("OTTO_SERVER")
	defer func() {
		serverURL = oldServerURL
		if oldEnvVar != "" {
			os.Setenv("OTTO_SERVER", oldEnvVar)
		} else {
			os.Unsetenv("OTTO_SERVER")
		}
	}()

	// Set via environment variable
	serverURL = ""
	os.Setenv("OTTO_SERVER", ts.URL)

	// Create output buffer
	output := new(bytes.Buffer)
	cmdOutput = output

	// Run the command
	cmd := &cobra.Command{}
	statsRun(cmd, []string{})

	// Verify we got remote stats
	if !bytes.Contains(output.Bytes(), []byte("Goroutines")) {
		t.Error("Expected output to contain Goroutines")
	}
}

// Helper function to ensure client package is imported
func init() {
	_ = client.NewClient("http://localhost:8011")
}
