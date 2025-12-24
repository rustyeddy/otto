package cmd

import (
	"io"
	"log/slog"
	"os"

	"github.com/rustyeddy/otto/client"
	"github.com/spf13/cobra"
)

var (
	appdir     string
	cmdOutput  io.Writer
	serverURL  string
	remoteMode bool
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
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "Otto server URL (e.g., http://localhost:8011)")
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

// GetClient returns an Otto client if remote mode is enabled, nil otherwise.
// It checks the --server flag first, then the OTTO_SERVER environment variable.
func GetClient() *client.Client {
	// Check if serverURL was set via --server flag
	if serverURL == "" {
		// Check environment variable
		serverURL = os.Getenv("OTTO_SERVER")
	}

	// If we have a server URL, create and return a client
	if serverURL != "" {
		return client.NewClient(serverURL)
	}

	return nil
}

// IsRemoteMode returns true if commands should connect to a remote server
func IsRemoteMode() bool {
	return GetClient() != nil
}
