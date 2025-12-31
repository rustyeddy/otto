// Demo program showing the new flexible logging configuration
package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"github.com/rustyeddy/otto/utils"
)

func main() {
	fmt.Println("=== Otto Logging Demo ===")
	fmt.Println()

	// Demo 1: Log to stdout with text format
	fmt.Println("1. Logging to stdout (text format):")
	config1 := utils.LogConfig{
		Level:  "info",
		Output: utils.LogOutputStdout,
		Format: utils.LogFormatText,
	}
	utils.InitLogger(config1)
	slog.Info("Application started", "version", "1.0.0", "mode", "demo")
	fmt.Println()

	// Demo 2: Log to stderr with JSON format
	fmt.Println("2. Logging to stderr (JSON format):")
	config2 := utils.LogConfig{
		Level:  "warn",
		Output: utils.LogOutputStderr,
		Format: utils.LogFormatJSON,
	}
	utils.InitLogger(config2)
	slog.Warn("System resource usage high", "cpu_percent", 85, "memory_mb", 2048)
	fmt.Println()

	// Demo 3: Log to string buffer
	fmt.Println("3. Logging to string buffer:")
	config3 := utils.LogConfig{
		Level:  "debug",
		Output: utils.LogOutputString,
		Format: utils.LogFormatText,
	}
	buffer, _ := utils.InitLogger(config3)
	slog.Debug("Debug message", "component", "auth", "user_id", 12345)
	slog.Info("Info message captured in buffer")
	fmt.Printf("   Buffer contents:\n   %s\n", buffer.String())

	// Demo 4: Log to string buffer with JSON format
	fmt.Println("4. Logging to string buffer (JSON format):")
	config4 := utils.LogConfig{
		Level:  "info",
		Output: utils.LogOutputString,
		Format: utils.LogFormatJSON,
	}
	jsonBuffer, _ := utils.InitLogger(config4)
	slog.Info("JSON formatted log", "event", "user_login", "timestamp", 1234567890)
	fmt.Printf("   JSON output:\n   %s\n", jsonBuffer.String())

	// Demo 5: Log to file
	fmt.Println("5. Logging to file (/tmp/otto-demo.log):")
	config5 := utils.LogConfig{
		Level:    "debug",
		Output:   utils.LogOutputFile,
		Format:   utils.LogFormatText,
		FilePath: "/tmp/otto-demo.log",
	}
	_, err := utils.InitLogger(config5)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		slog.Debug("Debug log written to file")
		slog.Info("Info log written to file")
		slog.Warn("Warning log written to file", "alert", "disk_space_low")

		// Read and display file contents
		content, _ := os.ReadFile("/tmp/otto-demo.log")
		fmt.Printf("   File contents:\n   %s\n", string(content))
	}

	// Demo 6: Using custom buffer
	fmt.Println("6. Using a custom buffer:")
	customBuffer := &bytes.Buffer{}
	config6 := utils.LogConfig{
		Level:  "info",
		Output: utils.LogOutputString,
		Format: utils.LogFormatJSON,
		Buffer: customBuffer,
	}
	utils.InitLogger(config6)
	slog.Info("Custom buffer log", "sensor", "temperature", "value", 23.5)
	fmt.Printf("   Custom buffer: %s\n", customBuffer.String())

	fmt.Println()
	fmt.Println("\n=== Demo Complete ===")
}
