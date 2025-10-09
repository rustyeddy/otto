package cmd

import (
	"bytes"
	"testing"

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
