package otto

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/rustyeddy/otto"
)

func TestFileCmdRegistration(t *testing.T) {
	// Test that fileCmd is registered with rootCmd
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "file" {
			found = true
			break
		}
	}
	if !found {
		t.Error("fileCmd should be registered with rootCmd")
	}
}

func TestFileRunSuccess(t *testing.T) {
	// Create a temporary file with test commands
	tmpFile, err := os.CreateTemp("", "otto_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test commands to the file
	testCommands := []string{
		"help",
		"version",
		"# This is a comment",
		"",
		"help",
	}
	for _, cmd := range testCommands {
		_, err := tmpFile.WriteString(cmd + "\n")
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}
	tmpFile.Close()

	// Track RunLine calls
	originalRunLine := RunLine
	var executedLines []string
	RunLine = func(line string) bool {
		executedLines = append(executedLines, line)
		return true
	}
	defer func() { RunLine = originalRunLine }()

	// Save original Interactive value
	originalInteractive := otto.Interactive
	defer func() { otto.Interactive = originalInteractive }()

	// Test fileRun
	args := []string{tmpFile.Name()}
	fileRun(fileCmd, args)

	// Verify Interactive is set
	if !otto.Interactive {
		t.Error("otto.Interactive should be set to true")
	}

	// Verify all lines were processed
	expectedLines := []string{"help", "version", "# This is a comment", "", "help"}
	if len(executedLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(executedLines))
	}

	for i, expected := range expectedLines {
		if i < len(executedLines) && executedLines[i] != expected {
			t.Errorf("Line %d: expected %q, got %q", i, expected, executedLines[i])
		}
	}
}

func TestFileRunFileNotFound(t *testing.T) {
	// Capture log output
	var logOutput strings.Builder
	originalHandler := slog.Default().Handler()
	logger := slog.New(slog.NewTextHandler(&logOutput, nil))
	slog.SetDefault(logger)
	defer slog.SetDefault(slog.New(originalHandler))

	// Test with non-existent file
	args := []string{"nonexistent_file.txt"}
	fileRun(fileCmd, args)

	// Check that error was logged
	logStr := logOutput.String()
	if !strings.Contains(logStr, "no such file") && !strings.Contains(logStr, "cannot find") {
		t.Error("Expected file not found error to be logged")
	}
}

func TestFileRunEmptyFile(t *testing.T) {
	// Create empty temporary file
	tmpFile, err := os.CreateTemp("", "otto_empty_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Track RunLine calls
	originalRunLine := RunLine
	var executedLines []string
	RunLine = func(line string) bool {
		executedLines = append(executedLines, line)
		return true
	}
	defer func() { RunLine = originalRunLine }()

	// Test fileRun with empty file
	args := []string{tmpFile.Name()}
	fileRun(fileCmd, args)

	// Verify no lines were processed
	if len(executedLines) != 0 {
		t.Errorf("Expected no lines to be processed, got %d", len(executedLines))
	}
}

func TestFileRunWithCommentsAndEmptyLines(t *testing.T) {
	// Create a temporary file with mixed content
	tmpFile, err := os.CreateTemp("", "otto_mixed_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write mixed content
	content := `# This is a comment
help

version
# Another comment

help`
	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Track RunLine calls
	originalRunLine := RunLine
	var executedLines []string
	RunLine = func(line string) bool {
		executedLines = append(executedLines, line)
		return true
	}
	defer func() { RunLine = originalRunLine }()

	// Test fileRun
	args := []string{tmpFile.Name()}
	fileRun(fileCmd, args)

	// Verify all lines were processed (including comments and empty lines)
	expectedLines := []string{
		"# This is a comment",
		"help",
		"",
		"version",
		"# Another comment",
		"",
		"help",
	}

	if len(executedLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(executedLines))
	}

	for i, expected := range expectedLines {
		if i < len(executedLines) && executedLines[i] != expected {
			t.Errorf("Line %d: expected %q, got %q", i, expected, executedLines[i])
		}
	}
}

func TestFileRunNoArguments(t *testing.T) {
	// Test with no arguments - should panic or handle gracefully
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when no arguments provided")
		}
	}()

	args := []string{}
	fileRun(fileCmd, args)
}

func TestFileRunPermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create a temporary file and remove read permissions
	tmpFile, err := os.CreateTemp("", "otto_perm_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Remove read permissions
	err = os.Chmod(tmpFile.Name(), 0000)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer os.Chmod(tmpFile.Name(), 0644) // Restore for cleanup

	// Capture log output
	var logOutput strings.Builder
	originalHandler := slog.Default().Handler()
	logger := slog.New(slog.NewTextHandler(&logOutput, nil))
	slog.SetDefault(logger)
	defer slog.SetDefault(slog.New(originalHandler))

	// Test fileRun
	args := []string{tmpFile.Name()}
	fileRun(fileCmd, args)

	// Check that permission error was logged
	logStr := logOutput.String()
	if !strings.Contains(logStr, "permission denied") {
		t.Error("Expected permission denied error to be logged")
	}
}

func TestFileRunInteractiveMode(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "otto_interactive_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("help\n")
	tmpFile.Close()

	// Save original Interactive value
	originalInteractive := otto.Interactive
	otto.Interactive = false
	defer func() { otto.Interactive = originalInteractive }()

	// Mock RunLine to avoid actual command execution
	originalRunLine := RunLine
	RunLine = func(line string) bool { return true }
	defer func() { RunLine = originalRunLine }()

	// Test fileRun
	args := []string{tmpFile.Name()}
	fileRun(fileCmd, args)

	// Verify Interactive is set to true
	if !otto.Interactive {
		t.Error("fileRun should set otto.Interactive to true")
	}
}

func TestFileRunLargeFile(t *testing.T) {
	// Create a large temporary file
	tmpFile, err := os.CreateTemp("", "otto_large_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write many lines
	const numLines = 1000
	for i := 0; i < numLines; i++ {
		_, err := tmpFile.WriteString(fmt.Sprintf("command%d\n", i))
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}
	tmpFile.Close()

	// Track RunLine calls
	originalRunLine := RunLine
	var executedCount int
	RunLine = func(line string) bool {
		executedCount++
		return true
	}
	defer func() { RunLine = originalRunLine }()

	// Test fileRun
	args := []string{tmpFile.Name()}
	fileRun(fileCmd, args)

	// Verify all lines were processed
	if executedCount != numLines {
		t.Errorf("Expected %d lines to be processed, got %d", numLines, executedCount)
	}
}
