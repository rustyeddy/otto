package cmd

import (
	"fmt"

	"github.com/rustyeddy/otto/utils"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display runtime stats",
	Long:  `Display runtime stats`,
	Run:   statsRun,
}

func statsRun(cmd *cobra.Command, args []string) {
	stats := utils.GetStats()
	fmt.Fprintf(cmdOutput, "Stats: %+v\n", stats)
}
