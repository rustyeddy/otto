package cmd

import (
	"fmt"
	"time"

	"github.com/rustyeddy/otto/station"
	"github.com/spf13/cobra"
)

var stationCmd = &cobra.Command{
	Use:   "station",
	Short: "Get station information",
	Long:  `Get a list of stations as well as details of a given station`,
	Run:   stationRun,
}

func stationRun(cmd *cobra.Command, args []string) {
	stations := station.GetStationManager()
	if stations == nil || stations.Count() == 0 {
		return
	}

	for _, st := range stations.Stations {
		fmt.Fprintf(cmdOutput, "station: %s: %s/%v\n",
			st.ID, st.LastHeard.Format(time.RFC3339), st.Expiration)
	}
}
