package cmd

import (
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	appdir    string
	cmdOutput io.Writer
)

var rootCmd = &cobra.Command{
	Use:   "otto",
	Short: "OttO is an IoT platform for creating cool IoT apps and hubs",
	Long: `This is cool stuff and you will be able to find a lot of cool information 
                in the following documentation https://rustyeddy.com/otto/`,
	Run: ottoRun,
}

func init() {
	cmdOutput = os.Stdout
	rootCmd.PersistentFlags().StringVar(&appdir, "appdir", "embed", "root of the web app")
	rootCmd.SetOut(cmdOutput)

	rootCmd.AddCommand(cliCmd)
	rootCmd.AddCommand(fileCmd)

	rootCmd.AddCommand(msgCmd)
	msgCmd.AddCommand(msgConnectCmd)
	msgCmd.AddCommand(msgPubCmd)
	msgCmd.AddCommand(msgSubCmd)

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(stationCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(tickerCmd)
	rootCmd.AddCommand(versionCmd)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		return
	}
}

func ottoRun(cmd *cobra.Command, args []string) {
	serveRun(cmd, args)
}
