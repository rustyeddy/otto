package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/rustyeddy/otto/client"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of otto",
		Long:  `All software has versions. This is OttO's`,
		Run: func(cmd *cobra.Command, args []string) {
			cli := client.NewClient("http://localhost:8011")
			version, err := cli.GetVersion()
			if err != nil {
				fmt.Fprintln(cmdOutput, "Failed to get otto client", err)
			}

			// Pretty print the JSON response
			jsonBytes, err := json.MarshalIndent(version, "", "  ")
			if err != nil {
				fmt.Fprintf(cmdOutput, "version: %+v\n", version)
				return
			}
			fmt.Fprintf(cmdOutput, "%s\n", string(jsonBytes))
		},
	}
)
