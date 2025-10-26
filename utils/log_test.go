package utils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupLogTest creates a temporary directory for test log files
func setupLogTest(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "otto-log-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tempDir
}

// teardownLogTest cleans up test files and directories
func teardownLogTest(t *testing.T, tempDir string) {
	err := os.RemoveAll(tempDir)
	if err != nil {
		t.Logf("Warning: failed to cleanup temp directory %s: %v", tempDir, err)
	}
}

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	io.Copy(&buf, r)
	return buf.String()
}

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "debug level",
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "info level",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level",
			input:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "error level",
			input:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "unknown level defaults to zero value (Info)",
			input:    "unknown",
			expected: slog.LevelInfo, // Zero value of slog.Level
		},
		{
			name:     "empty string defaults to zero value (Info)",
			input:    "",
			expected: slog.LevelInfo, // Zero value of slog.Level
		},
		{
			name:     "uppercase debug",
			input:    "DEBUG",
			expected: slog.LevelInfo, // should default since case-sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout for unknown level warnings
			var output string
			if tt.input == "unknown" || tt.input == "" || tt.input == "DEBUG" {
				output = captureStdout(func() {
					result := SetLogLevel(tt.input)
					if result != tt.expected {
						t.Errorf("SetLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
					}
				})

				// Check that warning message was printed for unknown levels
				if tt.input == "unknown" || tt.input == "DEBUG" {
					if !strings.Contains(output, "unknown loglevel") {
						t.Errorf("Expected warning message for unknown log level, got: %s", output)
					}
					if !strings.Contains(output, tt.input) {
						t.Errorf("Expected warning to mention input level %q, got: %s", tt.input, output)
					}
				}
			} else {
				result := SetLogLevel(tt.input)
				if result != tt.expected {
					t.Errorf("SetLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestInitLogger(t *testing.T) {
	tempDir := setupLogTest(t)
	defer teardownLogTest(t, tempDir)

	// Save original logfile value
	originalLogfile := logfile
	defer func() {
		logfile = originalLogfile
	}()

	// Test with custom log file
	testLogFile := filepath.Join(tempDir, "test.log")
	logfile = testLogFile

	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	t.Run("InitLogger with custom file and level", func(t *testing.T) {
		InitLogger("info", testLogFile)

		// Check that log file was created
		if _, err := os.Stat(testLogFile); os.IsNotExist(err) {
			t.Errorf("Expected log file %s to be created", testLogFile)
		}

		// Test that we can write to the logger
		slog.Info("test message")

		// Verify content was written to file
		content, err := os.ReadFile(testLogFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "test message") {
			t.Errorf("Expected log file to contain 'test message', got: %s", contentStr)
		}
		if !strings.Contains(contentStr, "INFO") {
			t.Errorf("Expected log file to contain 'INFO', got: %s", contentStr)
		}
	})

	t.Run("InitLogger with empty filename uses default", func(t *testing.T) {
		// Clean up previous test file
		os.Remove(testLogFile)

		InitLogger("debug", "")

		// Should use the default logfile (otto.log in tempDir)
		if _, err := os.Stat(testLogFile); os.IsNotExist(err) {
			t.Errorf("Expected default log file %s to be created", testLogFile)
		}
	})

	t.Run("InitLogger appends to existing file", func(t *testing.T) {
		// Create initial content
		InitLogger("warn", testLogFile)
		slog.Warn("first message")

		// Read initial content
		initialContent, err := os.ReadFile(testLogFile)
		if err != nil {
			t.Fatalf("Failed to read initial log content: %v", err)
		}

		// Wait a moment to ensure different timestamps
		time.Sleep(1 * time.Millisecond)

		// Initialize again (should append)
		InitLogger("error", testLogFile)
		slog.Error("second message")

		// Read final content
		finalContent, err := os.ReadFile(testLogFile)
		if err != nil {
			t.Fatalf("Failed to read final log content: %v", err)
		}

		finalStr := string(finalContent)
		if !strings.Contains(finalStr, "first message") {
			t.Error("Expected log file to contain original 'first message'")
		}
		if !strings.Contains(finalStr, "second message") {
			t.Error("Expected log file to contain new 'second message'")
		}
		if len(finalContent) <= len(initialContent) {
			t.Error("Expected log file to have grown after appending")
		}
	})
}

func TestInitLoggerWithDifferentLevels(t *testing.T) {
	tempDir := setupLogTest(t)
	defer teardownLogTest(t, tempDir)

	// Save original logfile value
	originalLogfile := logfile
	defer func() {
		logfile = originalLogfile
	}()

	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	testLogFile := filepath.Join(tempDir, "level-test.log")
	logfile = testLogFile

	tests := []struct {
		name      string
		level     string
		shouldLog map[string]bool // level -> should be logged
	}{
		{
			name:  "debug level logs everything",
			level: "debug",
			shouldLog: map[string]bool{
				"debug": true,
				"info":  true,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "info level logs info and above",
			level: "info",
			shouldLog: map[string]bool{
				"debug": false,
				"info":  true,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "warn level logs warn and above",
			level: "warn",
			shouldLog: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "error level logs only errors",
			level: "error",
			shouldLog: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  false,
				"error": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up previous test
			os.Remove(testLogFile)

			// Initialize logger with specific level
			InitLogger(tt.level, testLogFile)

			// Log messages at different levels
			slog.Debug("debug message")
			slog.Info("info message")
			slog.Warn("warn message")
			slog.Error("error message")

			// Read log content
			content, err := os.ReadFile(testLogFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			contentStr := string(content)

			// Check which messages should be present
			for level, shouldLog := range tt.shouldLog {
				message := fmt.Sprintf("%s message", level)
				contains := strings.Contains(contentStr, message)

				if shouldLog && !contains {
					t.Errorf("Expected log to contain '%s' but it didn't. Content: %s", message, contentStr)
				}
				if !shouldLog && contains {
					t.Errorf("Expected log NOT to contain '%s' but it did. Content: %s", message, contentStr)
				}
			}
		})
	}
}

func TestInitLoggerErrorHandling(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	// Save original logfile value
	originalLogfile := logfile
	defer func() {
		logfile = originalLogfile
	}()

	t.Run("InitLogger handles invalid directory", func(t *testing.T) {
		// Try to create log file in non-existent directory
		invalidPath := "/non/existent/directory/test.log"
		logfile = invalidPath

		// This should not panic, but will log an error
		// Capture the error output
		output := captureStdout(func() {
			InitLogger("info", invalidPath)
		})

		// The function should handle the error gracefully
		// We can't easily test the slog.Error output without more complex setup,
		// but we can ensure the function doesn't panic
		_ = output // Avoid unused variable warning
	})

	t.Run("InitLogger handles permission denied", func(t *testing.T) {
		// Create a directory with restricted permissions
		tempDir := setupLogTest(t)
		defer teardownLogTest(t, tempDir)

		restrictedDir := filepath.Join(tempDir, "restricted")
		err := os.Mkdir(restrictedDir, 0000) // No permissions
		if err != nil {
			t.Skipf("Could not create restricted directory: %v", err)
		}
		defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

		restrictedFile := filepath.Join(restrictedDir, "test.log")
		logfile = restrictedFile

		// This should handle the permission error gracefully
		InitLogger("info", restrictedFile)

		// The function should not panic, even if file creation fails
	})
}

// Benchmark tests
func BenchmarkSetLogLevel(b *testing.B) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}

	for i := 0; i < b.N; i++ {
		level := levels[i%len(levels)]
		SetLogLevel(level)
	}
}

func BenchmarkInitLogger(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "otto-log-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save original logfile value
	originalLogfile := logfile
	defer func() {
		logfile = originalLogfile
	}()

	testLogFile := filepath.Join(tempDir, "bench.log")
	logfile = testLogFile

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InitLogger("info", testLogFile)
	}
}

// Tests for new LogConfig functionality
func TestLogConfig_OutputToStdout(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "info",
		Output: LogOutputStdout,
		Format: LogFormatText,
	}

	_, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	// This should not panic or error
	slog.Info("test stdout message")
}

func TestLogConfig_OutputToStderr(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "warn",
		Output: LogOutputStderr,
		Format: LogFormatText,
	}

	_, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	// This should not panic or error
	slog.Warn("test stderr message")
}

func TestLogConfig_OutputToFile(t *testing.T) {
	tempDir := setupLogTest(t)
	defer teardownLogTest(t, tempDir)

	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	testLogFile := filepath.Join(tempDir, "config-test.log")

	config := LogConfig{
		Level:    "debug",
		Output:   LogOutputFile,
		Format:   LogFormatText,
		FilePath: testLogFile,
	}

	_, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	// Write some log messages
	slog.Debug("debug message")
	slog.Info("info message")
	slog.Warn("warn message")

	// Verify content was written to file
	content, err := os.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "debug message") {
		t.Errorf("Expected log file to contain 'debug message', got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "info message") {
		t.Errorf("Expected log file to contain 'info message', got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "warn message") {
		t.Errorf("Expected log file to contain 'warn message', got: %s", contentStr)
	}
}

func TestLogConfig_OutputToString(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "info",
		Output: LogOutputString,
		Format: LogFormatText,
	}

	buffer, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	if buffer == nil {
		t.Fatal("Expected buffer to be returned for string output")
	}

	// Write log messages
	slog.Info("test string message")
	slog.Warn("warning string message")

	// Check buffer content
	content := buffer.String()
	if !strings.Contains(content, "test string message") {
		t.Errorf("Expected buffer to contain 'test string message', got: %s", content)
	}
	if !strings.Contains(content, "warning string message") {
		t.Errorf("Expected buffer to contain 'warning string message', got: %s", content)
	}
}

func TestLogConfig_OutputToStringWithProvidedBuffer(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	providedBuffer := &bytes.Buffer{}
	config := LogConfig{
		Level:  "debug",
		Output: LogOutputString,
		Format: LogFormatText,
		Buffer: providedBuffer,
	}

	returnedBuffer, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	if returnedBuffer != providedBuffer {
		t.Error("Expected returned buffer to be the same as provided buffer")
	}

	// Write log messages
	slog.Debug("debug in custom buffer")

	// Check buffer content
	content := providedBuffer.String()
	if !strings.Contains(content, "debug in custom buffer") {
		t.Errorf("Expected buffer to contain 'debug in custom buffer', got: %s", content)
	}
}

func TestLogConfig_JSONFormat(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "info",
		Output: LogOutputString,
		Format: LogFormatJSON,
	}

	buffer, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	// Write log messages
	slog.Info("json test message", "key1", "value1", "key2", 42)

	// Check that output is valid JSON
	content := buffer.String()
	if !strings.Contains(content, `"msg":"json test message"`) {
		t.Errorf("Expected JSON format with msg field, got: %s", content)
	}
	if !strings.Contains(content, `"key1":"value1"`) {
		t.Errorf("Expected JSON format with key1 field, got: %s", content)
	}
	if !strings.Contains(content, `"key2":42`) {
		t.Errorf("Expected JSON format with key2 field, got: %s", content)
	}
	// Check for JSON structure indicators
	if !strings.Contains(content, "{") || !strings.Contains(content, "}") {
		t.Errorf("Expected JSON format with braces, got: %s", content)
	}
}

func TestLogConfig_TextFormat(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "info",
		Output: LogOutputString,
		Format: LogFormatText,
	}

	buffer, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig failed: %v", err)
	}

	// Write log messages
	slog.Info("text test message", "field1", "value1")

	// Check that output is in text format (not JSON)
	content := buffer.String()
	if !strings.Contains(content, "text test message") {
		t.Errorf("Expected text format with message, got: %s", content)
	}
	// Text format should have key=value pairs
	if !strings.Contains(content, "field1=value1") {
		t.Errorf("Expected text format with field1=value1, got: %s", content)
	}
}

func TestLogConfig_DifferentLevels(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	tests := []struct {
		name      string
		level     string
		logFunc   func(string)
		shouldLog bool
	}{
		{"debug logs at debug level", "debug", func(msg string) { slog.Debug(msg) }, true},
		{"info logs at debug level", "debug", func(msg string) { slog.Info(msg) }, true},
		{"debug does not log at info level", "info", func(msg string) { slog.Debug(msg) }, false},
		{"info logs at info level", "info", func(msg string) { slog.Info(msg) }, true},
		{"warn logs at warn level", "warn", func(msg string) { slog.Warn(msg) }, true},
		{"info does not log at warn level", "warn", func(msg string) { slog.Info(msg) }, false},
		{"error logs at error level", "error", func(msg string) { slog.Error(msg) }, true},
		{"warn does not log at error level", "error", func(msg string) { slog.Warn(msg) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LogConfig{
				Level:  tt.level,
				Output: LogOutputString,
				Format: LogFormatText,
			}

			buffer, err := InitLoggerWithConfig(config)
			if err != nil {
				t.Fatalf("InitLoggerWithConfig failed: %v", err)
			}

			testMsg := "test message for " + tt.name
			tt.logFunc(testMsg)

			content := buffer.String()
			contains := strings.Contains(content, testMsg)

			if tt.shouldLog && !contains {
				t.Errorf("Expected message to be logged, but it wasn't. Content: %s", content)
			}
			if !tt.shouldLog && contains {
				t.Errorf("Expected message NOT to be logged, but it was. Content: %s", content)
			}
		})
	}
}

func TestLogConfig_InvalidFilePathError(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:    "info",
		Output:   LogOutputFile,
		Format:   LogFormatText,
		FilePath: "/non/existent/directory/test.log",
	}

	_, err := InitLoggerWithConfig(config)
	if err == nil {
		t.Error("Expected error for invalid file path, got nil")
	}
}

func TestLogConfig_DefaultOutput(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "info",
		Output: LogOutput("invalid"),
		Format: LogFormatText,
	}

	// Should default to stdout and not error
	_, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig should not fail with invalid output type: %v", err)
	}
}

func TestLogConfig_DefaultFormat(t *testing.T) {
	// Store original default logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	config := LogConfig{
		Level:  "info",
		Output: LogOutputString,
		Format: LogFormat("invalid"),
	}

	buffer, err := InitLoggerWithConfig(config)
	if err != nil {
		t.Fatalf("InitLoggerWithConfig should not fail with invalid format: %v", err)
	}

	// Should default to text format
	slog.Info("test default format")
	content := buffer.String()

	// Text format uses key=value pairs
	if !strings.Contains(content, "test default format") {
		t.Errorf("Expected text format output, got: %s", content)
	}
}
