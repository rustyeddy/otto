package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "Get station information",
	Long:  `Get a list of stations as well as details of a given station`,
	Run:   stationsRun,
}

func stationsRun(cmd *cobra.Command, args []string) {
	client := GetClient()
	if client == nil {
		fmt.Fprintf(cmdOutput, "Failed to get an otto client")
		return
	}

	// Remote mode: fetch stations from server
	slog.Debug("Fetching stations from remote server", "url", client.BaseURL)
	stationsData, err := client.GetStations()
	if err != nil {
		fmt.Fprintf(cmdOutput, "Error fetching remote stations: %v\n", err)
		return
	}

	// Pretty print the JSON response
	jsonBytes, err := json.MarshalIndent(stationsData, "", "  ")
	if err != nil {
		fmt.Fprintf(cmdOutput, "Stations: %+v\n", stationsData)
		return
	}
	fmt.Fprintf(cmdOutput, "%s\n", string(jsonBytes))
}
