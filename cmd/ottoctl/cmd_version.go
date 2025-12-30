package ottoctl

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	shutdownCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of otto",
		Long:  `All software has versions. This is OttO's`,
		Args:  cobra.MaximumNArgs(0),
		RunE:  runVersion,
	}
)

func runVersion(cmd *cobra.Command, args []string) error {

	cli := getClient()
	vmap, err := cli.GetVersion()
	if err != nil {
		fmt.Fprintln(cmdOutput, "Failed to get otto client", err)
	}
	version, ex := vmap["version"]
	if !ex {
		fmt.Fprintf(errOutput, "%s\n", "failed to get version")
		return errors.New("failed to get version")
	}

	fmt.Fprintf(cmdOutput, "%s\n", version)
	return nil
}
