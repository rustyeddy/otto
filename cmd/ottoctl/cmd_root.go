package ottoctl

import (
	"io"
	"log/slog"
	"os"

	"github.com/rustyeddy/otto/client"
	"github.com/spf13/cobra"
)

var (
	cmdOutput io.Writer
	errOutput io.Writer
	serverURL string
	format    string
	cli       *client.Client
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
	errOutput = os.Stderr
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8011", "Otto server URL (e.g., http://localhost:8011)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "choices are <human | json> default json")
	rootCmd.SetOut(cmdOutput)
	rootCmd.SetErr(errOutput)

	rootCmd.AddCommand(cliCmd)
	rootCmd.AddCommand(stationsCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(versionCmd)
}

func getClient() *client.Client {
	if cli == nil {
		cli = GetClient()
	}
	return cli
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
}

// GetClient returns an Otto client if remote mode is enabled, nil otherwise.
// It checks the --server flag first, then the OTTO_SERVER environment variable.
func GetClient() *client.Client {
	// Start with the value provided via --server flag (if any).
	effectiveURL := serverURL

	// If no flag was provided, fall back to the environment variable.
	if effectiveURL == "" {
		if envURL := os.Getenv("OTTO_SERVER"); envURL != "" {
			effectiveURL = envURL
		}
	}

	// If we have a server URL, create and return a client.
	if effectiveURL != "" {
		return client.NewClient(effectiveURL)
	}

	return nil
}

// IsRemoteMode returns true if commands should connect to a remote server
func IsRemoteMode() bool {
	return GetClient() != nil
}
