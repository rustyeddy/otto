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

// get is a helper method that performs a GET request and decodes JSON response.
// It handles common error cases and reduces code duplication.
func (c *Client) get(path string, result interface{}) error {
	resp, err := c.HTTPClient.Get(c.BaseURL + path)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// GetStats retrieves runtime statistics from the Otto server.
// Returns a Stats struct containing goroutine count, CPU info, memory stats, etc.
//
// This calls the /api/stats endpoint on the server.
func (c *Client) GetStats() (map[string]interface{}, error) {
	var stats map[string]interface{}
	if err := c.get("/api/stats", &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetStations retrieves a list of all stations from the Otto server.
// Returns a map containing stations and stale station information.
//
// This calls the /api/stations endpoint on the server.
func (c *Client) GetStations() (map[string]interface{}, error) {
	var stations map[string]interface{}
	if err := c.get("/api/stations", &stations); err != nil {
		return nil, err
	}
	return stations, nil
}

// GetVersion retrieves version number of the server.
//
// This calls the /version endpoint on the server.
func (c *Client) GetVersion() (map[string]interface{}, error) {
	var ver map[string]interface{}

	if err := c.get("/version", &ver); err != nil {
		return nil, err
	}
	return ver, nil
}

// GetVersion retrieves version number of the server.
//
// This calls the /version endpoint on the server.
func (c *Client) Shutdown() (map[string]any, error) {
	var result map[string]interface{}

	if err := c.get("/api/shutdown", &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetLogConfig retrieves the log configuration
func (c *Client) GetLogConfig() (map[string]any, error) {
	var result map[string]any

	if err := c.get("/api/log", &result); err != nil {
		return result, err
	}
	return result, nil

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
