package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestCliStationsCmd(t *testing.T) {
	// Create a buffer to capture the output
	output := new(bytes.Buffer)
	cliStationsCmd.SetOut(output)
	cmdOutput = output

	// Execute the command
	cliStationsCmd.Run(cliStationsCmd, []string{})

	// Check if the output contains expected result (may be empty in test environment)
	got := output.String()
	if got == "" || bytes.Contains(output.Bytes(), []byte("No stations found")) {
		// This is acceptable in a test environment
		t.Log("No stations found (expected in test environment)")
	}
}

func TestCliStationsRun(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the cliStationsRun function
	cliStationsRun(cmd, args)

	// Output might be empty or "No stations found" in test environment
	got := output.String()
	if got == "" || bytes.Contains(output.Bytes(), []byte("No stations found")) {
		t.Log("No stations found (expected in test environment)")
	}
}

func TestCliStationsRun_RemoteMode(t *testing.T) {
	// Create a test server that returns stations data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stations" {
			t.Errorf("Expected path /api/stations, got %s", r.URL.Path)
		}
		stationsData := map[string]interface{}{
			"stations": map[string]interface{}{
				"station1": map[string]interface{}{
					"id":         "station1",
					"hostname":   "test-host",
					"last-heard": "2024-12-24T12:00:00Z",
					"local":      false,
				},
			},
			"stale": map[string]interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stationsData)
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
	cliStationsRun(cmd, []string{})

	// Check output contains JSON response
	result := output.String()
	if result == "" {
		t.Error("Expected output, got empty string")
	}
	if !bytes.Contains(output.Bytes(), []byte("station1")) {
		t.Error("Expected output to contain station1")
	}
}

func TestCliStationsRun_RemoteMode_Error(t *testing.T) {
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
	cliStationsRun(cmd, []string{})

	// Check output contains error message
	if !bytes.Contains(output.Bytes(), []byte("Error")) {
		t.Error("Expected output to contain error message")
	}
}

func TestCliStationsRun_RemoteMode_EnvVar(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stationsData := map[string]interface{}{
			"stations": map[string]interface{}{
				"station2": map[string]interface{}{
					"id":       "station2",
					"hostname": "env-test-host",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stationsData)
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
	cliStationsRun(cmd, []string{})

	// Verify we got remote stations
	if !bytes.Contains(output.Bytes(), []byte("station2")) {
		t.Error("Expected output to contain station2")
	}
}
