// serve test

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestServeCmd(t *testing.T) {
	cmd := serveCmd

	if cmd.Use != "serve" {
		t.Errorf("expected Use to be 'serve', got '%s'", cmd.Use)
	}

	if cmd.Short != "Start oTTo the Server" {
		t.Errorf("expected Short to be 'Start oTTo the Server', got '%s'", cmd.Short)
	}

	if cmd.Long != "Start OttO the IoT Server" {
		t.Errorf("expected Long to be 'Start OttO the IoT Server', got '%s'", cmd.Long)
	}

	if cmd.Run == nil {
		t.Error("expected Run to be set, got nil")
	}
}

func TestServeCmdFlags(t *testing.T) {
	cmd := serveCmd

	flag := cmd.Flags().Lookup("foreground")
	if flag == nil {
		t.Error("expected 'foreground' flag to be defined, got nil")
	}

	if flag.Usage != "Run the server command in the foreground" {
		t.Errorf("expected 'foreground' flag usage to be 'Run the server command in the foreground', got '%s'", flag.Usage)
	}
}

func TestServeRun(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}

	// Call serveRun to ensure it doesn't panic or throw errors
	serveRun(cmd, args)
}
