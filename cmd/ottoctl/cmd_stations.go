package ottoctl

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "Get station information",
	Long:  `Get a list of stations as well as details of a given station`,
	RunE:  stationsRun,
}

func stationsRun(cmd *cobra.Command, args []string) error {
	client := getClient()
	if client == nil {
		return errors.New("Failed to get an otto client")
	}

	// Remote mode: fetch stations from server
	slog.Debug("Fetching stations from remote server", "url", client.BaseURL)
	stationsData, err := client.GetStations()
	if err != nil {
		fmt.Fprintln(errOutput, err)
		return errors.New(fmt.Sprintf("Error fetching remote stations: %v\n", err))
	}

	// // Pretty print the JSON response
	// jsonBytes, err := json.MarshalIndent(stationsData, "", "  ")
	// if err != nil {
	// 	fmt.Fprintf(cmdOutput, "Stations: %+v\n", stationsData)
	// 	return
	// }
	fmt.Fprintf(cmdOutput, "ID Hostname		LastHeard\n")
	fmt.Fprintf(cmdOutput, "-------------------------\n")
	for _, st := range stationsData {
		fmt.Fprintf(cmdOutput, "%s: %s %d\n", st.ID, st.Hostname, st.LastHeard)
	}
	return nil
}
