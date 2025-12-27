package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display runtime stats",
	Long:  `Display runtime stats from local or remote Otto instance`,
	Run:   statsRun,
}

func statsRun(cmd *cobra.Command, args []string) {
	// Check if we should connect to a remote server
	client := GetClient()
	if client == nil {
		fmt.Fprintf(cmdOutput, "Failed to get an otto client")
		return
	}

	// Remote mode: fetch stats from server
	slog.Debug("Fetching stats from remote server", "url", client.BaseURL)
	stats, err := client.GetStats()
	if err != nil {
		fmt.Fprintf(cmdOutput, "Error fetching remote stats: %v\n", err)
		return
	}

	// Pretty print the JSON response
	jsonBytes, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Fprintf(cmdOutput, "Stats: %+v\n", stats)
	} else {
		fmt.Fprintf(cmdOutput, "%s\n", string(jsonBytes))
	}
}
