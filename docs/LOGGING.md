# Otto Logging Configuration

The Otto IoT framework provides flexible logging configuration using Go's standard `log/slog` package with support for multiple output destinations and formats.

## Features

- **Multiple Output Destinations**: Write logs to file, stdout, stderr, or in-memory string buffer
- **Multiple Formats**: Choose between human-readable text format or machine-readable JSON format
- **Configurable Log Levels**: Set logging level to debug, info, warn, or error
- **Backward Compatible**: Legacy `InitLogger()` function still supported
- **Structured Logging**: Uses slog for structured, typed logging with key-value pairs

## Quick Start

### Basic Usage with LogConfig

```go
import (
    "log/slog"
    "github.com/rustyeddy/otto/utils"
)

// Configure logger to write to stdout in text format
config := utils.LogConfig{
    Level:  "info",
    Output: utils.LogOutputStdout,
    Format: utils.LogFormatText,
}

utils.InitLoggerWithConfig(config)
slog.Info("Application started", "version", "1.0.0")
```

### Log to File

```go
config := utils.LogConfig{
    Level:    "debug",
    Output:   utils.LogOutputFile,
    Format:   utils.LogFormatText,
    FilePath: "/var/log/otto.log",
}

_, err := utils.InitLoggerWithConfig(config)
if err != nil {
    fmt.Printf("Failed to initialize logger: %v\n", err)
}

slog.Debug("Detailed debug information")
slog.Info("Application started")
```

### Log to String Buffer (for testing)

```go
config := utils.LogConfig{
    Level:  "info",
    Output: utils.LogOutputString,
    Format: utils.LogFormatText,
}

buffer, _ := utils.InitLoggerWithConfig(config)
slog.Info("Test message", "key", "value")

// Access log content
logContent := buffer.String()
```

### JSON Format Logging

```go
config := utils.LogConfig{
    Level:  "info",
    Output: utils.LogOutputStdout,
    Format: utils.LogFormatJSON,
}

utils.InitLoggerWithConfig(config)
slog.Info("User logged in", "user_id", 12345, "ip", "192.168.1.1")
// Output: {"time":"2025-10-26T12:00:00Z","level":"INFO","msg":"User logged in","user_id":12345,"ip":"192.168.1.1"}
```

### Log to stderr

```go
config := utils.LogConfig{
    Level:  "warn",
    Output: utils.LogOutputStderr,
    Format: utils.LogFormatText,
}

utils.InitLoggerWithConfig(config)
slog.Warn("High memory usage", "memory_mb", 2048)
```

### Custom Buffer

```go
import "bytes"

customBuffer := &bytes.Buffer{}
config := utils.LogConfig{
    Level:  "debug",
    Output: utils.LogOutputString,
    Format: utils.LogFormatJSON,
    Buffer: customBuffer,
}

utils.InitLoggerWithConfig(config)
slog.Debug("Custom buffer log")

// Use your custom buffer
fmt.Printf("Logged %d bytes\n", customBuffer.Len())
```

## LogConfig Structure

```go
type LogConfig struct {
    Level      string         // Log level: "debug", "info", "warn", "error"
    Output     LogOutput      // Output destination
    Format     LogFormat      // Log format
    FilePath   string         // File path (when Output is LogOutputFile)
    Buffer     *bytes.Buffer  // Custom buffer (when Output is LogOutputString)
}
```

### Output Options

- `LogOutputFile` - Write logs to a file
- `LogOutputStdout` - Write logs to standard output
- `LogOutputStderr` - Write logs to standard error
- `LogOutputString` - Write logs to an in-memory buffer

### Format Options

- `LogFormatText` - Human-readable text format with key=value pairs
- `LogFormatJSON` - Machine-readable JSON format

### Log Levels

- `"debug"` - Most verbose, includes all logs
- `"info"` - Informational messages and above
- `"warn"` - Warnings and errors only
- `"error"` - Error messages only

## Backward Compatibility

The legacy `InitLogger()` function is still supported:

```go
// Old API still works
utils.InitLogger("info", "/var/log/otto.log")
slog.Info("Using legacy API")
```

## Examples

See the [examples/logging_demo.go](examples/logging_demo.go) file for comprehensive examples demonstrating all features.

Run the demo:
```bash
go run examples/logging_demo.go
```

## Best Practices

1. **Production**: Use file output with log rotation, JSON format for parsing
   ```go
   config := utils.LogConfig{
       Level:    "info",
       Output:   utils.LogOutputFile,
       Format:   utils.LogFormatJSON,
       FilePath: "/var/log/otto.log",
   }
   ```

2. **Development**: Use stdout with text format for readability
   ```go
   config := utils.LogConfig{
       Level:  "debug",
       Output: utils.LogOutputStdout,
       Format: utils.LogFormatText,
   }
   ```

3. **Testing**: Use string buffer to capture and assert log content
   ```go
   config := utils.LogConfig{
       Level:  "debug",
       Output: utils.LogOutputString,
       Format: utils.LogFormatText,
   }
   buffer, _ := utils.InitLoggerWithConfig(config)
   // ... run tests ...
   assert.Contains(t, buffer.String(), "expected log message")
   ```

4. **Always use structured logging** with key-value pairs:
   ```go
   // Good
   slog.Info("User action", "action", "login", "user_id", 123)
   
   // Avoid
   slog.Info(fmt.Sprintf("User %d performed login", 123))
   ```

## Testing

The logging package includes comprehensive tests:

```bash
go test ./utils -v -run "TestLogConfig"
```

All tests:
```bash
go test ./utils -v
```
