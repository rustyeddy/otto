package ottoctl

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "shutdown",
		Short: "Shutdown the otto server",
		Long:  `Shutdown the otto server nicely allowing it to cleanup after itself`,
		Args:  cobra.MaximumNArgs(0),
		Run:   runShutdown,
	}
)

func runShutdown(cmd *cobra.Command, args []string) {
	cli := getClient()
	result, err := cli.Shutdown()
	if err != nil {
		fmt.Fprintln(cmdOutput, "Failed to get otto client", err)
		return
	}
	fmt.Fprintf(cmdOutput, "%s\n", result["shutdown"])
	return
}
