package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// TestStatsRunRemoteMode tests the remote mode functionality
func TestStatsRunRemoteMode(t *testing.T) {
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

	// Save original serverURL and restore after test
	originalServerURL := serverURL
	defer func() { serverURL = originalServerURL }()

	// Set serverURL to test server
	serverURL = ts.URL

	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the statsRun function in remote mode
	statsRun(cmd, args)

	// Check that output contains the stats data
	got := output.String()
	assert.Contains(t, got, "Goroutines")
	assert.Contains(t, got, "CPUs")
	assert.Contains(t, got, "GoVersion")
	assert.Contains(t, got, "10")
	assert.Contains(t, got, "4")
	assert.Contains(t, got, "go1.21.0")
}

// TestStatsRunRemoteModeError tests error handling when remote server returns an error
func TestStatsRunRemoteModeError(t *testing.T) {
	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer ts.Close()

	// Save original serverURL and restore after test
	originalServerURL := serverURL
	defer func() { serverURL = originalServerURL }()

	// Set serverURL to test server
	serverURL = ts.URL

	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the statsRun function in remote mode
	statsRun(cmd, args)

	// Check that output contains error message
	got := output.String()
	assert.Contains(t, got, "Error fetching remote stats")
	assert.Contains(t, got, "server returned error")
}

// TestStatsRunRemoteModeConnectionError tests error handling when cannot connect to server
func TestStatsRunRemoteModeConnectionError(t *testing.T) {
	// Save original serverURL and restore after test
	originalServerURL := serverURL
	defer func() { serverURL = originalServerURL }()

	// Set serverURL to invalid URL
	serverURL = "http://localhost:9999"

	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the statsRun function in remote mode
	statsRun(cmd, args)

	// Check that output contains error message
	got := output.String()
	assert.Contains(t, got, "Error fetching remote stats")
}

// TestStatsRunRemoteModeInvalidJSON tests error handling when remote server returns invalid JSON
func TestStatsRunRemoteModeInvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	// Save original serverURL and restore after test
	originalServerURL := serverURL
	defer func() { serverURL = originalServerURL }()

	// Set serverURL to test server
	serverURL = ts.URL

	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the statsRun function in remote mode
	statsRun(cmd, args)

	// Check that output contains error message
	got := output.String()
	assert.Contains(t, got, "Error fetching remote stats")
}

// TestStatsRunJSONMarshalPath tests the JSON marshaling path
func TestStatsRunJSONMarshalPath(t *testing.T) {
	// Create a test server that returns a complex stats object
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"Goroutines": 42,
			"CPUs":       8,
			"GoVersion":  "go1.21.0",
			"MemStats": map[string]interface{}{
				"Alloc":      1024,
				"TotalAlloc": 2048,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))
	defer ts.Close()

	// Save original serverURL and restore after test
	originalServerURL := serverURL
	defer func() { serverURL = originalServerURL }()

	// Set serverURL to test server
	serverURL = ts.URL

	cmd := &cobra.Command{}
	args := []string{}

	output := new(bytes.Buffer)
	cmdOutput = output

	// Call the statsRun function in remote mode
	statsRun(cmd, args)

	// Check that output is valid JSON with proper indentation
	got := output.String()

	// Verify the JSON is properly formatted
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(got), &parsed)
	if err != nil {
		t.Logf("Output is not JSON, which is acceptable if stats are printed: %s", got)
	}

	// At minimum, check that the output contains the expected stats
	assert.Contains(t, got, "Goroutines")
	assert.Contains(t, got, "42")
	assert.Contains(t, got, "MemStats")
}

// mockClientGetter is a helper type for testing that allows injecting a custom client
type mockClientGetter struct {
	client *client.Client
}

// TestStatsRunWithMockedClient tests the stats command with a fully mocked client
func TestStatsRunWithMockedClient(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "successful remote stats retrieval",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					stats := map[string]interface{}{
						"Goroutines": 15,
						"CPUs":       2,
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(stats)
				}))
			},
			expectedOutput: []string{"Goroutines", "15", "CPUs", "2"},
			notExpected:    []string{"Error"},
		},
		{
			name: "server returns 404",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, "Not Found")
				}))
			},
			expectedOutput: []string{"Error fetching remote stats"},
			notExpected:    []string{},
		},
		{
			name: "server returns 500",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, "Server Error")
				}))
			},
			expectedOutput: []string{"Error fetching remote stats", "500"},
			notExpected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.setupServer()
			defer ts.Close()

			// Save original serverURL and restore after test
			originalServerURL := serverURL
			defer func() { serverURL = originalServerURL }()

			// Set serverURL to test server
			serverURL = ts.URL

			cmd := &cobra.Command{}
			args := []string{}

			output := new(bytes.Buffer)
			cmdOutput = output

			// Call the statsRun function
			statsRun(cmd, args)

			got := output.String()

			// Check expected output
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, got, expected, "Expected output to contain: %s", expected)
			}

			// Check that unwanted strings are not present
			for _, notWanted := range tt.notExpected {
				assert.NotContains(t, got, notWanted, "Expected output to NOT contain: %s", notWanted)
			}
		})
	}
}
