package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestGetRootCmd(t *testing.T) {
	cmd := GetRootCmd()
	if cmd == nil {
		t.Fatal("expected rootCmd to be non-nil")
	}

	if cmd.Use != "otto" {
		t.Errorf("expected Use to be 'otto', got '%s'", cmd.Use)
	}

	if cmd.Short != "OttO is an IoT platform for creating cool IoT apps and hubs" {
		t.Errorf("unexpected Short description: %s", cmd.Short)
	}
}

func TestExecute(t *testing.T) {
	// Replace the default rootCmd with a mock command for testing
	oldRoot := rootCmd
	defer func() { rootCmd = oldRoot }()
	mockCmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			// Mock behavior
		},
	}
	rootCmd = mockCmd

	// Execute the command and ensure no errors occur
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOttoRun(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}

	// Call ottoRun to ensure it doesn't panic or throw errors
	// (now calls serveRun instead of showing usage)
	go ottoRun(cmd, args)
	time.Sleep(1 * time.Second)

}

func TestGetClient_NoServerURL(t *testing.T) {
	// Save original values and restore after test
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Clear environment variable
	oldEnv := os.Getenv("OTTO_SERVER")
	os.Unsetenv("OTTO_SERVER")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OTTO_SERVER", oldEnv)
		}
	}()

	// Clear serverURL flag
	serverURL = ""

	client := GetClient()
	if client != nil {
		t.Error("Expected GetClient() to return nil when no server URL is provided")
	}
}

func TestGetClient_WithServerFlag(t *testing.T) {
	// Save original values and restore after test
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Clear environment variable
	oldEnv := os.Getenv("OTTO_SERVER")
	os.Unsetenv("OTTO_SERVER")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OTTO_SERVER", oldEnv)
		}
	}()

	// Set serverURL flag
	serverURL = "http://localhost:8011"

	client := GetClient()
	if client == nil {
		t.Fatal("Expected GetClient() to return a client when --server flag is set")
	}
	if client.BaseURL != "http://localhost:8011" {
		t.Errorf("Expected client BaseURL to be 'http://localhost:8011', got '%s'", client.BaseURL)
	}
}

func TestGetClient_WithEnvVar(t *testing.T) {
	// Save original values and restore after test
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Set environment variable
	oldEnv := os.Getenv("OTTO_SERVER")
	os.Setenv("OTTO_SERVER", "http://envserver:9000")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OTTO_SERVER", oldEnv)
		} else {
			os.Unsetenv("OTTO_SERVER")
		}
	}()

	// Clear serverURL flag
	serverURL = ""

	client := GetClient()
	if client == nil {
		t.Fatal("Expected GetClient() to return a client when OTTO_SERVER env var is set")
	}
	if client.BaseURL != "http://envserver:9000" {
		t.Errorf("Expected client BaseURL to be 'http://envserver:9000', got '%s'", client.BaseURL)
	}
}

func TestGetClient_FlagOverridesEnvVar(t *testing.T) {
	// Save original values and restore after test
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Set both flag and environment variable
	serverURL = "http://flagserver:8011"

	oldEnv := os.Getenv("OTTO_SERVER")
	os.Setenv("OTTO_SERVER", "http://envserver:9000")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OTTO_SERVER", oldEnv)
		} else {
			os.Unsetenv("OTTO_SERVER")
		}
	}()

	client := GetClient()
	if client == nil {
		t.Fatal("Expected GetClient() to return a client")
	}
	if client.BaseURL != "http://flagserver:8011" {
		t.Errorf("Expected client BaseURL to be 'http://flagserver:8011' (from flag), got '%s'", client.BaseURL)
	}
}

func TestIsRemoteMode_WithClient(t *testing.T) {
	// Save original values and restore after test
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Clear environment variable
	oldEnv := os.Getenv("OTTO_SERVER")
	os.Unsetenv("OTTO_SERVER")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OTTO_SERVER", oldEnv)
		}
	}()

	// Set serverURL flag
	serverURL = "http://localhost:8011"

	if !IsRemoteMode() {
		t.Error("Expected IsRemoteMode() to return true when server URL is set")
	}
}

func TestIsRemoteMode_WithoutClient(t *testing.T) {
	// Save original values and restore after test
	oldServerURL := serverURL
	defer func() { serverURL = oldServerURL }()

	// Clear environment variable
	oldEnv := os.Getenv("OTTO_SERVER")
	os.Unsetenv("OTTO_SERVER")
	defer func() {
		if oldEnv != "" {
			os.Setenv("OTTO_SERVER", oldEnv)
		}
	}()

	// Clear serverURL flag
	serverURL = ""

	if IsRemoteMode() {
		t.Error("Expected IsRemoteMode() to return false when no server URL is set")
	}
}
