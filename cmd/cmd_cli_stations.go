package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/rustyeddy/otto/station"
	"github.com/spf13/cobra"
)

var cliStationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "Display stations",
	Long:  `Display stations from local or remote Otto instance`,
	Run:   cliStationsRun,
}

func cliStationsRun(cmd *cobra.Command, args []string) {
	// Check if we should connect to a remote server
	client := GetClient()
	if client == nil {
		// Local mode: get stations directly
		slog.Debug("Getting stations from local process")
		stations := station.GetStationManager()
		if stations == nil || stations.Count() == 0 {
			fmt.Fprintf(cmdOutput, "No stations found\n")
			return
		}

		// Print stations in a simple format
		for id, st := range stations.Stations {
			fmt.Fprintf(cmdOutput, "Station: %s\n", id)
			fmt.Fprintf(cmdOutput, "  Hostname: %s\n", st.Hostname)
			fmt.Fprintf(cmdOutput, "  Last Heard: %s\n", st.LastHeard.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(cmdOutput, "  Local: %v\n", st.Local)
			if len(st.Ifaces) > 0 {
				fmt.Fprintf(cmdOutput, "  Interfaces:\n")
				for _, iface := range st.Ifaces {
					fmt.Fprintf(cmdOutput, "    - %s (%s)\n", iface.Name, iface.MACAddr)
					for _, ip := range iface.IPAddrs {
						fmt.Fprintf(cmdOutput, "      %s\n", ip.String())
					}
				}
			}
			fmt.Fprintf(cmdOutput, "\n")
		}
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
	} else {
		fmt.Fprintf(cmdOutput, "%s\n", string(jsonBytes))
	}
}
