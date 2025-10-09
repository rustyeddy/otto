package cmd

import (
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start oTTo the Server",
	Long:  `Start OttO the IoT Server`,
	Run:   serveRun,
}

var (
	foreground bool
)

func init() {
	serveCmd.Flags().BoolVar(&foreground, "foreground", false, "Run the server command in the foreground")
}

func serveRun(cmd *cobra.Command, args []string) {
	// otto.OttO()
}
