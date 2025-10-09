package cmd

import (
	"bytes"
	"testing"

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

	// Capture the output of cmd.Usage()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	ottoRun(cmd, args)

	output := buf.String()
	if output == "" {
		t.Error("expected usage output, got empty string")
	}
}
