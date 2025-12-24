# Client-Server Architecture

This document describes the client-server architecture implemented in Otto, allowing commands to connect to a running Otto server instance.

## Overview

Otto commands can now operate in two modes:
- **Local Mode**: Commands execute directly on the local machine, accessing local resources
- **Remote Mode**: Commands connect to a running Otto server via REST API

## Architecture Components

### 1. Client Package (`client/`)
The client package provides a REST client for connecting to Otto servers:
- `Client` struct with methods for API calls
- `GetStats()` - Retrieve runtime statistics
- `Ping()` - Health check

### 2. Server REST API (`server/`)
New REST API endpoints exposed by Otto server:
- `/api/stats` - Runtime statistics (goroutines, memory, CPU)
- `/api/stations` - Station information (planned)
- `/api/timers` - Timer information (planned)
- `/ping` - Health check endpoint

### 3. Command Updates (`cmd/`)
Commands updated to support both local and remote modes:
- Auto-detect mode based on `--server` flag or `OTTO_SERVER` env var
- Seamless switching between local and remote execution

## Usage

### Starting the Server

```bash
# Start Otto server (listens on port 8011 by default)
otto serve
```

### Local Mode (Default)

```bash
# Run commands locally without connecting to a server
otto stats
```

### Remote Mode - Using Flag

```bash
# Connect to a specific server
otto stats --server http://localhost:8011
```

### Remote Mode - Using Environment Variable

```bash
# Set the server URL via environment variable
export OTTO_SERVER=http://localhost:8011
otto stats
```

## Example Output

### Local Mode
```
Stats: &{Goroutines:1 CPUs:2 MemStats:{...} GoVersion:go1.25.4}
```

### Remote Mode
```json
{
  "Alloc": 756264,
  "CPUs": 2,
  "GoVersion": "go1.25.4",
  "Goroutines": 11,
  "HeapAlloc": 756264,
  ...
}
```

## Architecture Diagram

```
┌─────────────────┐
│  CLI Commands   │
│  (Remote Mode)  │
└────────┬────────┘
         │
         v
┌────────────────┐
│  REST Client   │
│  (client pkg)  │
└────────┬───────┘
         │ HTTP/JSON
         v
┌────────────────┐
│  Otto Server   │
│    :8011       │
├────────────────┤
│ /api/stats     │
│ /api/stations  │
│ /api/timers    │
│ /ping          │
└────────────────┘
```

## Benefits

1. **Separation of Concerns**: CLI commands can query a running server without needing direct access to resources
2. **Remote Management**: Manage Otto instances from different machines
3. **Scalability**: Multiple clients can connect to a single server
4. **Monitoring**: Query runtime statistics without disrupting the server
5. **Testing**: Easy to test commands against test servers

## Next Steps

Commands to convert to client-server architecture:
- [x] `stats` - Runtime statistics (COMPLETED)
- [ ] `station` - Station information
- [ ] `timers` - Timer information
- [ ] `msg pub` - Publish messages via MQTT
- [ ] `msg sub` - Subscribe to messages (WebSocket)

## Implementation Details

### Detection Logic
```go
func GetClient() *client.Client {
    // Check --server flag first
    if serverURL != "" {
        return client.NewClient(serverURL)
    }
    
    // Check OTTO_SERVER environment variable
    if url := os.Getenv("OTTO_SERVER"); url != "" {
        return client.NewClient(url)
    }
    
    return nil // Local mode
}
```

### Command Pattern
```go
func commandRun(cmd *cobra.Command, args []string) {
    if client := GetClient(); client != nil {
        // Remote mode: call REST API
        data, err := client.GetData()
        // ... handle response
    } else {
        // Local mode: direct access
        data := localPackage.GetData()
        // ... handle local data
    }
}
```

## Testing

```bash
# Build
make build

# Run tests
make test

# Test client package specifically
go test ./client/... -v

# Test server endpoints
go test ./server/... -v

# Manual testing
./otto serve &
./otto stats --server http://localhost:8011
pkill -f "otto serve"
```
