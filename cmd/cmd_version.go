package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of otto",
		Long:  `All software has versions. This is OttO's`,
		Run: func(cmd *cobra.Command, args []string) {
			cli := client.GetClient()
			version, err := cli.Get("/version")
			println("VERSION -----------------", version)
			fmt.Fprintln(cmdOutput, version)
		},
	}
)
