package otto

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	version = "0.1.0"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of otto",
		Long:  `All software has versions. This is OttO's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
)
