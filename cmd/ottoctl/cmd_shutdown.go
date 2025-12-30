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
		RunE:  runShutdown,
	}
)

func runShutdown(cmd *cobra.Command, args []string) error {

	cli := getClient()
	err := cli.Shutdown()
	if err != nil {
		fmt.Fprintln(cmdOutput, "Failed to get otto client", err)
		return err
	}
	fmt.Fprintf(cmdOutput, "shutdown request sent")
	return nil
}
