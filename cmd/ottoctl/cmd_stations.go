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

	fmt.Fprintf(cmdOutput, "%20s: %-20s %8s %9s %16s\n", "id", "hostname", "expires", "version", "ipaddrs")
	fmt.Fprintf(cmdOutput, "-----------------------------------------------------------------------------------\n")
	for _, st := range stationsData {
		//rounded := st.LastHeard.Round(time.Second)
		exp := st.LastHeard.Add(st.Expiration)
		expires := time.Until(exp)
		expires = expires.Round(time.Second)
		ipaddr := ""
		if len(st.Ifaces) > 0 && len(st.Ifaces[0].IPAddrs) > 0 {
			ipaddr = st.Ifaces[0].IPAddrs[0].String()
		}
		fmt.Fprintf(cmdOutput, "%20s: %-20s %8v %9s %16s\n", st.ID, st.Hostname, expires, st.Version, ipaddr)
	}
	return nil
}
