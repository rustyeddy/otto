package otto

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

func init() {
	rootCmd.AddCommand(tickerCmd)
}

func tickerRun(cmd *cobra.Command, args []string) {
	t := utils.GetTickers()
	fmt.Printf("%+v\n", t)
}
