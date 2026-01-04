package ottoctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var timersCmd = &cobra.Command{
	Use:   "timers",
	Short: "Display timer information",
	Long:  `Display timer information from local or remote Otto instance`,
	RunE:  timersRun,
}

func timersRun(cmd *cobra.Command, args []string) error {
	// Check if we should connect to a remote server
	client := getClient()
	if client == nil {
		return errors.New("timersRun failed to get a client")
	}

	// Remote mode: fetch timers from server
	slog.Debug("Fetching timers from remote server", "url", client.BaseURL)
	timers, err := client.GetTimers()
	if err != nil {
		fmt.Fprintf(errOutput, "Error fetching remote timers: %v\n", err)
		return err
	}

	// Pretty print the JSON response
	jsonBytes, err := json.MarshalIndent(timers, "", "  ")
	if err != nil {
		fmt.Fprintf(errOutput, "%s", err)
		return err
	} else {
		fmt.Fprintf(cmdOutput, "%s\n", string(jsonBytes))
	}
	return nil
}
