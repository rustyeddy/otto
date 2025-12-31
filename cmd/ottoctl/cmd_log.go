package ottoctl

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	logCmd = &cobra.Command{
		Use:   "log",
		Short: "Display and configure logging",
		Long:  `Display and configure logging`,
		RunE:  runLog,
	}
)

func runLog(cmd *cobra.Command, args []string) error {
	cli := getClient()
	lc, err := cli.GetLogConfig()
	if err != nil {
		fmt.Fprintln(cmdOutput, "otto client failed to retrieve log config", err)
		return err
	}
	fmt.Fprintf(cmdOutput, "Output: %s\n", lc.Output)
	fmt.Fprintf(cmdOutput, "Format: %s\n", lc.Format)
	fmt.Fprintf(cmdOutput, "FilePath: %s\n", lc.FilePath)
	fmt.Fprintf(cmdOutput, "Buffer: %s\n", lc.Buffer)
	return nil
}
