package utils_test

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/rustyeddy/otto/utils"
)

// Example_logToStdout demonstrates logging to stdout
func Example_logToStdout() {
	config := utils.LogConfig{
		Level:  "info",
		Output: utils.LogOutputStdout,
		Format: utils.LogFormatText,
	}

	utils.InitLogger(config)
	slog.Info("Application started successfully")
	// Output will go to stdout
}

// Example_logToFile demonstrates logging to a file
func Example_logToFile() {
	config := utils.LogConfig{
		Level:    "debug",
		Output:   utils.LogOutputFile,
		Format:   utils.LogFormatText,
		FilePath: "/tmp/app.log",
	}

	_, err := utils.InitLogger(config)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	slog.Debug("Debug information")
	slog.Info("Application started")
	slog.Warn("Warning message")
}

// Example_logToString demonstrates logging to a string buffer
func Example_logToString() {
	config := utils.LogConfig{
		Level:  "info",
		Output: utils.LogOutputString,
		Format: utils.LogFormatText,
	}

	buffer, err := utils.InitLogger(config)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	slog.Info("Test message", "key", "value")

	// Access the log content from buffer
	logContent := buffer.String()
	fmt.Printf("Logged: %s", logContent)
}

// Example_logJSONFormat demonstrates JSON formatted logging
func Example_logJSONFormat() {
	config := utils.LogConfig{
		Level:  "info",
		Output: utils.LogOutputStderr,
		Format: utils.LogFormatJSON,
	}

	utils.InitLogger(config)
	slog.Info("User logged in", "user_id", 12345, "ip", "192.168.1.1")
	// Output will be in JSON format to stderr
}

// Example_customBuffer demonstrates using a custom buffer
func Example_customBuffer() {
	customBuffer := &bytes.Buffer{}
	config := utils.LogConfig{
		Level:  "debug",
		Output: utils.LogOutputString,
		Format: utils.LogFormatJSON,
		Buffer: customBuffer,
	}

	utils.InitLogger(config)
	slog.Debug("Debug with custom buffer", "component", "auth")

	// Use the custom buffer
	fmt.Printf("Buffer contains %d bytes\n", customBuffer.Len())
}
