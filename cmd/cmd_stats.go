package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/rustyeddy/otto/utils"
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
		// Local mode: get stats directly
		slog.Debug("Getting stats from local process")
		stats := utils.GetStats()
		fmt.Fprintf(cmdOutput, "Stats: %+v\n", stats)
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
