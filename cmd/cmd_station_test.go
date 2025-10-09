package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStationCmd(t *testing.T) {
	// Create a buffer to capture the output
	output := new(bytes.Buffer)
	rootCmd.SetOut(output)

	// Add the station command to the root command
	rootCmd.AddCommand(stationCmd)

	// Execute the station command
	// err := stationCmd.RunE(&cobra.Command{}, []string{})
	stationCmd.Run(rootCmd, []string{})

	// Check if the output contains expected content
	result := output.String()
	assert.Empty(t, result)
}
