# OttO - IoT Device Framework

OttO is a Go framework for building IoT applications with clean separation between hardware device interaction and messaging infrastructure. It provides a flexible, testable architecture for sensor stations, garden automation, and other IoT projects.

## Recent Architectural Improvements ✨

### 🏗️ **Clean Device/Messaging Architecture**
- **Separation of Concerns**: Device layer focuses purely on hardware interaction, Otto handles IoT messaging infrastructure
- **ManagedDevice Wrapper**: Bridges any simple device with messaging capabilities (MQTT pub/sub, event handling)
- **Type-Safe Device Management**: Re-enabled Station device management with proper interfaces
- **Reusable Components**: Any device from the `devices` package can be easily wrapped with messaging

### 🌐 **Flexible Messaging Options**
- **Local Messaging**: Internal pub/sub for testing and development (no external dependencies)
- **MQTT with Fallback**: Attempts MQTT connection, gracefully falls back to local messaging
- **Public MQTT Support**: Default integration with `test.mosquitto.org` for easy testing
- **Custom MQTT Brokers**: Configurable broker URLs for production deployments

### ✅ **Production Ready Features**
- **Mock Mode**: Complete hardware abstraction for development and testing
- **Web Interface**: Full-featured UI for monitoring and control
- **RESTful API**: Standard HTTP endpoints for integration
- **Robust Error Handling**: Graceful degradation when hardware/network unavailable 

## Quick Start

### Development/Testing (Mock Mode)
```bash
# Local messaging, no hardware required
./your-app -mock -local

# MQTT with public test broker
./your-app -mock

# Custom MQTT broker
./your-app -mock -mqtt-broker your.broker.com
```

### Hardware Deployment
```bash
# Production mode with real sensors
./your-app

# With custom MQTT broker
./your-app -mqtt-broker your.production.broker.com
```

## Usage Examples

### Garden Station Implementation
The `garden-station` project demonstrates the full architecture:
- **Sensors**: BME280 (temperature/humidity/pressure), VH400 (soil moisture)
- **Actuators**: Water pump, LED indicators, OLED display
- **Controls**: Physical buttons for manual override
- **Automation**: Automatic watering based on soil moisture thresholds
- **Web UI**: Real-time monitoring and manual controls at http://localhost:8011

### Command Line Options
- `-mock`: Enable hardware mocking for development
- `-local`: Force local messaging (no MQTT)
- `-mqtt-broker string`: Custom MQTT broker address

## Architecture Overview

### Device Layer (`devices` package)
```go
// Simple, focused device interfaces
type Device[T any] interface {
    ID() string
    Type() Type
    Open() error
    Close() error
    Get() (T, error)
    Set(v T) error
}
```

### Messaging Layer (`otto` package)
```go
// ManagedDevice wraps devices with messaging
type ManagedDevice struct {
    Name   string
    Device any
    Topic  string
    messanger.Messanger
}
```

# Messaging Infrastructure

### MQTT Broker 

- Run MQTT broker, e.g. mosquitto

- Base topic "ss/<id>/data/<data-type>"

Example: ```ss/00:95:fb:3f:34:95/data/tempc 25.00```

### Web Sockets

We sockets or HTTP/2 will be used to send data to and from the IOTe
device (otto) in our case.

### Subscribe to Topics

- announce/station  - announces stations that control or collect
- announce/hub      - announces hubs, typ

- data/tempc/       - data can have option /index at the end
- data/humidity

- control/relay/idx - control can have option /index at the end

## REST API

- GET   /api/config 
- PUT   /api/config     data => { config: id, ... }

- GET /api/data
- GET /api/stations

## Station Manager 

- Collection of stations
- Stations can age out 

### Stations

- ID (name, IP and mac address)
- Capabilities
  - sensors
  - relay

## Data

Data can be optimized and we expect we will want to optimize different
data for all kinds of reasons and we won't preclude that from
happening, we'll give applications the flexibility to handle data
elements as they see fit (can optimize).

We will take an memory expensive approach, every data point can be
handled on it's own. The data structure will be:

    struct Data
        Source ID
        Type
        Timestamp
        Value

# User Interface


# Build

1. Install Go 
2. go get ./...
3. cd ss; go build 

That should leave the execuable 'sensors' in the 'sensors' directory as so:

> ./station/sensors/sensors

## Deploy

1. Install and run an MQTT broker on the sensors host
(e.g. mosquitto).

2. Start the _sensors_ program ensuring the sensor station has
connected to a wifi network.

3. Put batteries in sensors and let the network build itself.

## Testing & Development

### Mock Mode Testing
```bash
# Complete hardware abstraction - no GPIO/I2C devices needed
./garden-station -mock -local

# Test with public MQTT broker
./garden-station -mock

# Web interface available at http://localhost:8011
curl http://localhost:8011/
```

### Integration Testing
- **Local Messaging**: Zero external dependencies, instant startup
- **MQTT Testing**: Uses `test.mosquitto.org` by default
- **Device Mocking**: All hardware interactions simulated
- **Web UI Testing**: Full-featured interface for manual testing

## Current Status ✅

### ✅ **Completed Features:**
- Clean device/messaging architecture implemented
- ManagedDevice wrapper for any device type
- Flexible messaging (local + MQTT with fallback)
- Complete hardware mocking support
- Web interface fully functional
- MQTT connectivity with public/custom brokers
- Type-safe device management
- Garden station reference implementation

### 🚀 **Ready For:**
1. **Hardware Deployment**: Remove `-mock` flag and connect real sensors
2. **Production MQTT**: Connect to existing MQTT infrastructure  
3. **Custom Applications**: Use as framework for new IoT projects
4. **Scaling**: Add more device types and stations

### 📊 **Testing Results:**
- Mock mode: ✅ All GPIO/I2C errors eliminated
- Web interface: ✅ Full garden station UI functional
- MQTT connectivity: ✅ Successfully connects to test.mosquitto.org
- Local messaging: ✅ Clean fallback when MQTT unavailable
- Device abstraction: ✅ Hardware interactions properly mocked

## Building & Deployment

### Prerequisites
- Go 1.21+ 
- For hardware: Linux with GPIO/I2C support (Raspberry Pi recommended)

### Build
```bash
git clone https://github.com/rustyeddy/otto
cd otto
go mod tidy
go build ./...
```

### Deploy
```bash
# Development
./your-app -mock -local

# Production  
./your-app -mqtt-broker your.broker.com
```
