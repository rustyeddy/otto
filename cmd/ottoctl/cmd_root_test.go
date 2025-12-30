package ottoctl

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

func TestGetClient(t *testing.T) {
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

	t.Run("returns nil when no server URL is provided", func(t *testing.T) {
		serverURL = ""
		os.Unsetenv("OTTO_SERVER")
		client := getClient()
		if client != nil {
			t.Error("Expected nil client when no server URL provided")
		}
		cli = nil
	})

	t.Run("returns client when --server flag is set", func(t *testing.T) {
		serverURL = "http://localhost:8011"
		os.Unsetenv("OTTO_SERVER")
		client := getClient()
		if client == nil {
			t.Fatal("Expected client when server URL is set")
		}
		if client.BaseURL != "http://localhost:8011" {
			t.Errorf("Expected BaseURL to be http://localhost:8011, got %s", client.BaseURL)
		}
		cli = nil
	})

	t.Run("returns client when OTTO_SERVER env var is set", func(t *testing.T) {
		serverURL = ""
		os.Setenv("OTTO_SERVER", "http://remote:8011")
		client := getClient()
		if client == nil {
			t.Fatal("Expected client when OTTO_SERVER env var is set")
		}
		if client.BaseURL != "http://remote:8011" {
			t.Errorf("Expected BaseURL to be http://remote:8011, got %s", client.BaseURL)
		}
		cli = nil
	})

	t.Run("prioritizes --server flag over env var", func(t *testing.T) {
		serverURL = "http://flag:8011"
		os.Setenv("OTTO_SERVER", "http://env:8011")
		client := getClient()
		if client == nil {
			t.Fatal("Expected client when both flag and env var are set")
		}
		if client.BaseURL != "http://flag:8011" {
			t.Errorf("Expected BaseURL to be http://flag:8011, got %s", client.BaseURL)
		}
		cli = nil
	})
}

func TestIsRemoteMode(t *testing.T) {
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
}
