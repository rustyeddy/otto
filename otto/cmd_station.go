package otto

import (
	"fmt"
	"time"

	"github.com/rustyeddy/otto/station"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stationCmd)
}

var stationCmd = &cobra.Command{
	Use:   "station",
	Short: "Get station information",
	Long:  `Get a list of stations as well as details of a given station`,
	Run:   stationRun,
}

func stationRun(cmd *cobra.Command, args []string) {
	stations := station.GetStationManager()
	for _, st := range stations.Stations {
		fmt.Printf("station: %s: %s/%v\n",
			st.ID, st.LastHeard.Format(time.RFC3339), st.Expiration)
	}
}
