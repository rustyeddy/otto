package ottoctl

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

var timersCmd = &cobra.Command{
	Use:   "timers",
	Short: "Display timer information",
	Long:  `Display timer information from local or remote Otto instance`,
	RunE:  timersRun,
}

func timersRun(cmd *cobra.Command, args []string) error {
	// Check if we should connect to a remote server
	client := getClient()
	if client == nil {
		return errors.New("timersRun failed to get a client")
	}

	// Remote mode: fetch timers from server
	slog.Debug("Fetching timers from remote server", "url", client.BaseURL)
	timers, err := client.GetTimers()
	if err != nil {
		fmt.Fprintf(errOutput, "Error fetching remote timers: %v\n", err)
		return err
	}

	fmt.Fprintf(cmdOutput, "%20s %7s %6s %s\n", "ticker", "active", "ticks", "last")
	fmt.Fprintln(cmdOutput, "--------------------------------------------------------------")
	for _, t := range timers {
		last := time.Since(t.LastTick)
		last = last.Round(time.Second)
		fmt.Fprintf(cmdOutput, "%20s %7t %6d %s\n", t.Name, t.Active, t.Ticks, last)
	}
	return nil
}
