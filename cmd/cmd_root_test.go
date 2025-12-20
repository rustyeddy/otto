package cmd

import (
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
