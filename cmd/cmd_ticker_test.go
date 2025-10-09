package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/rustyeddy/otto/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestTickerRun(t *testing.T) {
	// Mock the command output
	output := new(bytes.Buffer)

	// Create a mock command
	cmd := &cobra.Command{}
	cmd.SetOut(cmdOutput)
	cmdOutput = output

	utils.NewTicker("Timer1", time.Second, func(t time.Time) {})
	utils.NewTicker("Timer2", time.Second, func(t time.Time) {})

	// Run the tickerRun function
	tickerRun(cmd, []string{})

	// Verify the output
	assert.Contains(t, output.String(), "Timer1")
}
