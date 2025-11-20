package cmd

import (
	"fmt"

	"github.com/rustyeddy/otto/utils"
	"github.com/spf13/cobra"
)

var (
	tickerCmd = &cobra.Command{
		Use:   "timers",
		Short: "Display and manage timers",
		Long:  "Display and manage timers, stop, start and reset timers ",
		Run:   tickerRun,
	}
)

func tickerRun(cmd *cobra.Command, args []string) {
	tickers := utils.GetTickers()
	tstr := ""
	for n := range tickers {
		tstr += " " + n
	}
	fmt.Fprintf(cmdOutput, "%s\n", tstr)
}
