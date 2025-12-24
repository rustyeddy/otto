// Package client provides a client library for connecting to a remote Otto server.
// It supports REST API calls for querying server state and managing resources.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a connection to a remote Otto server.
// It provides methods for making REST API calls to the server.
type Client struct {
	// BaseURL is the base URL of the Otto server (e.g., "http://localhost:8011")
	BaseURL string

	// HTTPClient is the underlying HTTP client used for requests
	HTTPClient *http.Client
}

// NewClient creates a new Otto client connected to the specified server URL.
// The serverURL should include the protocol and port (e.g., "http://localhost:8011").
//
// Example:
//
//	client := client.NewClient("http://localhost:8011")
//	stats, err := client.GetStats()
func NewClient(serverURL string) *Client {
	return &Client{
		BaseURL: serverURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetStats retrieves runtime statistics from the Otto server.
// Returns a Stats struct containing goroutine count, CPU info, memory stats, etc.
//
// This calls the /api/stats endpoint on the server.
func (c *Client) GetStats() (map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/stats")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(body))
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stats, nil
}

// GetStations retrieves the list of stations from the Otto server.
// This calls the /api/stations endpoint on the server.
func (c *Client) GetStations() ([]map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/stations")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(body))
	}

	var stations []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stations, nil
}

// GetTimers retrieves the list of timers from the Otto server.
// This calls the /api/timers endpoint on the server.
func (c *Client) GetTimers() (map[string]interface{}, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/timers")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(body))
	}

	var timers map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&timers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return timers, nil
}

// Ping checks if the Otto server is reachable and responding.
// Returns nil if the server is healthy, error otherwise.
func (c *Client) Ping() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/ping")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %d", resp.StatusCode)
	}

	return nil
}
