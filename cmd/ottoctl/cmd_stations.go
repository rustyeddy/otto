package ottoctl

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

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

	fmt.Fprintf(cmdOutput, "ID Hostname		LastHeard\n")
	fmt.Fprintf(cmdOutput, "-------------------------\n")
	for _, st := range stationsData {
		rounded := st.LastHeard.Round(time.Second)
		fmt.Fprintf(cmdOutput, "%s: %s %s\n", st.ID, st.Hostname, rounded)
	}
	return nil
}
