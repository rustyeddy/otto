package otto

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock command for testing
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test command executed")
	},
}

var testSubCmd = &cobra.Command{
	Use:   "sub",
	Short: "Test subcommand",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test sub command executed")
	},
}

func TestCliCmdRegistration(t *testing.T) {
	t.Run("CLI command is registered", func(t *testing.T) {
		// Verify cliCmd is not nil
		assert.NotNil(t, cliCmd, "cliCmd should be initialized")
		assert.Equal(t, "cli", cliCmd.Use, "cliCmd should have correct Use field")
		assert.Equal(t, "Run auto in interactive CLI mode", cliCmd.Short, "cliCmd should have correct Short description")
		assert.NotNil(t, cliCmd.Run, "cliCmd should have Run function set")
	})

	t.Run("CLI command is added to root", func(t *testing.T) {
		// Check if cliCmd is in rootCmd's commands
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd == cliCmd {
				found = true
				break
			}
		}
		assert.True(t, found, "cliCmd should be added to rootCmd")
	})
}

func TestInitReadline(t *testing.T) {
	// Clean up any existing readline instance
	if rl != nil {
		rl.Close()
		rl = nil
	}

	t.Run("Initialize readline successfully", func(t *testing.T) {
		// Add test commands to rootCmd for testing completion
		rootCmd.AddCommand(testCmd)
		testCmd.AddCommand(testSubCmd)
		defer func() {
			rootCmd.RemoveCommand(testCmd)
		}()

		assert.NotPanics(t, func() {
			init_readline()
		}, "init_readline should not panic")

		assert.NotNil(t, rl, "readline instance should be created")

		// Clean up
		if rl != nil {
			rl.Close()
			rl = nil
		}
	})

	t.Run("Readline configuration", func(t *testing.T) {
		init_readline()
		defer func() {
			if rl != nil {
				rl.Close()
				rl = nil
			}
		}()

		assert.NotNil(t, rl, "readline instance should be created")
		// Note: We can't easily test the internal configuration of readline
		// but we can verify it was created without panicking
	})
}

func TestPcFromCommands(t *testing.T) {
	t.Run("Build completion tree", func(t *testing.T) {
		completer := readline.NewPrefixCompleter()

		// Create test command structure
		parentCmd := &cobra.Command{Use: "parent"}
		childCmd := &cobra.Command{Use: "child"}
		grandChildCmd := &cobra.Command{Use: "grandchild"}

		childCmd.AddCommand(grandChildCmd)
		parentCmd.AddCommand(childCmd)

		// Test building completion tree
		assert.NotPanics(t, func() {
			pcFromCommands(completer, parentCmd)
		}, "pcFromCommands should not panic")

		// Verify children were added
		children := completer.GetChildren()
		assert.NotEmpty(t, children, "Completer should have children")
	})

	t.Run("Handle command with no subcommands", func(t *testing.T) {
		completer := readline.NewPrefixCompleter()
		simpleCmd := &cobra.Command{Use: "simple"}

		assert.NotPanics(t, func() {
			pcFromCommands(completer, simpleCmd)
		}, "pcFromCommands should handle commands with no subcommands")
	})

	t.Run("Handle nil command", func(t *testing.T) {
		completer := readline.NewPrefixCompleter()

		// This would likely panic in real usage, but test defensive programming
		assert.Panics(t, func() {
			pcFromCommands(completer, nil)
		}, "pcFromCommands should panic with nil command")
	})
}

func TestRunLine(t *testing.T) {
	// Capture stdout for testing command execution
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	t.Run("Exit commands", func(t *testing.T) {
		testCases := []string{"exit", "quit", "  exit  ", "  quit  "}
		for _, input := range testCases {
			result := RunLine(input)
			assert.False(t, result, "RunLine should return false for exit command: %q", input)
		}
	})

	t.Run("Empty input", func(t *testing.T) {
		testCases := []string{"", "   ", "\t", "\n"}
		for _, input := range testCases {
			result := RunLine(input)
			assert.True(t, result, "RunLine should return true for empty input: %q", input)
		}
	})

	t.Run("Valid command execution", func(t *testing.T) {
		// Add test command
		rootCmd.AddCommand(testCmd)
		defer rootCmd.RemoveCommand(testCmd)

		// Capture output
		r, w, _ := os.Pipe()
		os.Stdout = w

		result := RunLine("test")

		w.Close()
		os.Stdout = originalStdout

		output := make([]byte, 1024)
		n, _ := r.Read(output)
		r.Close()

		assert.True(t, result, "RunLine should return true for valid command")
		assert.Contains(t, string(output[:n]), "test command executed", "Command should be executed")
	})

	t.Run("Invalid command", func(t *testing.T) {
		// Capture stderr for error message
		r, w, _ := os.Pipe()
		originalStderr := os.Stderr
		os.Stderr = w

		result := RunLine("nonexistent-command")

		w.Close()
		os.Stderr = originalStderr

		output := make([]byte, 1024)
		r.Read(output)
		r.Close()

		assert.True(t, result, "RunLine should return true even for invalid commands")
		assert.Contains(t, string(output), "Usage", "Usage should be displayed")

		// Note: Error handling might write to stdout instead of stderr
	})

	t.Run("Command with arguments", func(t *testing.T) {
		// Add test command that accepts args
		argTestCmd := &cobra.Command{
			Use: "argtest",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("argtest executed with %d args\n", len(args))
			},
		}
		rootCmd.AddCommand(argTestCmd)
		defer rootCmd.RemoveCommand(argTestCmd)

		// Capture output
		r, w, _ := os.Pipe()
		os.Stdout = w

		result := RunLine("argtest arg1 arg2")

		w.Close()
		os.Stdout = originalStdout

		output := make([]byte, 1024)
		n, _ := r.Read(output)
		r.Close()

		assert.True(t, result, "RunLine should return true for command with args")
		assert.Contains(t, string(output[:n]), "argtest executed", "Command with args should be executed")
	})

	t.Run("Command with flags", func(t *testing.T) {
		flagTestCmd := &cobra.Command{
			Use: "flagtest",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("flagtest executed")
			},
		}
		flagTestCmd.Flags().String("flag", "", "test flag")
		rootCmd.AddCommand(flagTestCmd)
		defer rootCmd.RemoveCommand(flagTestCmd)

		result := RunLine("flagtest --flag=value")
		assert.True(t, result, "RunLine should handle commands with flags")
	})
}

func TestCliLine(t *testing.T) {
	// Note: Testing cliLine is challenging because it depends on readline.Readline()
	// which reads from stdin. We'll test the logic we can control.

	t.Run("RunLine integration", func(t *testing.T) {
		// Test that cliLine calls RunLine with the correct behavior
		// This is more of an integration test since we can't easily mock readline

		// We can test the error handling logic by examining the function structure
		assert.NotNil(t, cliLine, "cliLine function should exist")
	})
}

func TestCliEdgeCases(t *testing.T) {
	t.Run("Multiple spaces in command", func(t *testing.T) {
		result := RunLine("   exit   ")
		assert.False(t, result, "Should handle multiple spaces around exit command")
	})

	t.Run("Mixed case exit commands", func(t *testing.T) {
		// The current implementation is case-sensitive
		result := RunLine("EXIT")
		assert.True(t, result, "Should be case-sensitive for exit commands")
	})

	t.Run("Command with special characters", func(t *testing.T) {
		result := RunLine("test-command-with-dashes")
		assert.True(t, result, "Should handle commands with special characters")
	})

	t.Run("Very long command line", func(t *testing.T) {
		longCmd := strings.Repeat("a", 1000)
		result := RunLine(longCmd)
		assert.True(t, result, "Should handle very long command lines")
	})

	t.Run("Command with quotes", func(t *testing.T) {
		result := RunLine(`test "argument with spaces"`)
		assert.True(t, result, "Should handle commands with quoted arguments")
	})
}

func TestCliRunFunction(t *testing.T) {
	t.Run("CliRun structure", func(t *testing.T) {
		// We can't easily test the full cliRun function since it's interactive
		// but we can verify it exists and has the right signature
		assert.NotNil(t, cliRun, "cliRun function should exist")

		// Verify it's set as the Run function for cliCmd
		assert.Equal(t, fmt.Sprintf("%p", cliRun), fmt.Sprintf("%p", cliCmd.Run),
			"cliCmd.Run should be set to cliRun")
	})
}

func TestReadlineCleanup(t *testing.T) {
	t.Run("Readline cleanup", func(t *testing.T) {
		// Initialize readline
		init_readline()
		require.NotNil(t, rl, "readline should be initialized")

		// Test cleanup
		assert.NotPanics(t, func() {
			rl.Close()
		}, "readline Close should not panic")

		// Reset for other tests
		rl = nil
	})
}

func TestCommandParsing(t *testing.T) {
	t.Run("Single word command", func(t *testing.T) {
		line := "help"
		args := strings.Split(line, " ")
		assert.Equal(t, []string{"help"}, args, "Single word should split correctly")
	})

	t.Run("Multi word command", func(t *testing.T) {
		line := "test arg1 arg2"
		args := strings.Split(line, " ")
		assert.Equal(t, []string{"test", "arg1", "arg2"}, args, "Multi word should split correctly")
	})

	t.Run("Command with extra spaces", func(t *testing.T) {
		line := "test  arg1   arg2"
		args := strings.Split(line, " ")
		// Note: strings.Split preserves empty strings between multiple spaces
		assert.Contains(t, args, "test", "Should contain command")
		assert.Contains(t, args, "arg1", "Should contain first arg")
		assert.Contains(t, args, "arg2", "Should contain second arg")
	})
}

// Benchmark tests for performance
func BenchmarkRunLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunLine("help")
	}
}

func BenchmarkRunLineExit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunLine("exit")
	}
}

func BenchmarkRunLineEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunLine("")
	}
}
