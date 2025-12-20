package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Run auto in interactive CLI mode",
	Long:  `Run auto in interactive CLI mode`,
	Run:   cliRun,
}

var (
	rl *readline.Instance
)

func init_readline() {
	var completer = readline.NewPrefixCompleter()
	for _, child := range rootCmd.Commands() {
		pcFromCommands(completer, child)
	}

	var err error
	rl, err = readline.NewEx(&readline.Config{
		Prompt:       "otto\033[31m»\033[0m ",
		HistoryFile:  "/tmp/readline.tmp",
		AutoComplete: completer,
		// InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		// FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	rl.CaptureExitSignal()
	// l.SetOutput(rl.Stderr())
}

func cliRun(cmd *cobra.Command, args []string) {

	init_readline()

	defer rl.Close()
	running := true
	for running {
		running = cliLine()
	}
	fmt.Fprintln(cmdOutput, "Exiting, cleanup")
	fmt.Fprintln(cmdOutput, "Good Bye!")
}

func pcFromCommands(parent readline.PrefixCompleterInterface, c *cobra.Command) {
	pc := readline.PcItem(c.Use)
	parent.SetChildren(append(parent.GetChildren(), pc))
	for _, child := range c.Commands() {
		pcFromCommands(pc, child)
	}
}

func cliLine() bool {
	line, err := rl.Readline()
	switch err {
	case readline.ErrInterrupt:
		if len(line) == 0 {
			return false
		} else {
			return true
		}
	case io.EOF:
		return false
	}

	return RunLine(line)
}

var RunLine = func(line string) bool {
	// func RunLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "exit" || line == "quit" {
		return false
	}

	if len(line) == 0 {
		return true
	}

	args := strings.Split(line, " ")
	cmd, args, err := rootCmd.Find(args)
	if err != nil {
		fmt.Fprintf(cmdOutput, "Error running cmd %q: %s\n", line, err)
	}

	cmd.ParseFlags(args)
	cmd.Run(cmd, args)
	return true
}
