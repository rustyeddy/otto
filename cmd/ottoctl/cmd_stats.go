package ottoctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display runtime stats",
	Long:  `Display runtime stats from local or remote Otto instance`,
	RunE:  statsRun,
}

func statsRun(cmd *cobra.Command, args []string) error {
	// Check if we should connect to a remote server
	client := getClient()
	if client == nil {
		return errors.New("statusRun failed to get a client")
	}

	// Remote mode: fetch stats from server
	slog.Debug("Fetching stats from remote server", "url", client.BaseURL)
	stats, err := client.GetStats()
	if err != nil {
		fmt.Fprintf(errOutput, "Error fetching remote stats: %v\n", err)
		return err
	}

	// Pretty print the JSON response
	jsonBytes, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Fprintf(errOutput, "%s", err)
		return err
	} else {
		fmt.Fprintf(cmdOutput, "%s\n", string(jsonBytes))
	}
	return nil
}
