package cmd

import (
	"bufio"
	"log/slog"
	"os"

	"github.com/rustyeddy/otto"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "read otto commands from a file",
	Long:  `Run otto with the commands in the file`,
	Run:   fileRun,
}

func fileRun(cmd *cobra.Command, args []string) {
	otto.Interactive = true
	fname := args[0]
	file, err := os.Open(fname)
	if err != nil {
		slog.Error(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		RunLine(line)
	}
	if err := scanner.Err(); err != nil {
		slog.Error(err.Error())
	}
}
