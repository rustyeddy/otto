# Building and Running Otto in Podman

## Quick Start

### Build the Image
```bash
podman build -t otto:latest .
```

### Run the Container

#### Basic (Mock Mode with Local Messaging)
```bash
podman run -p 8011:8011 otto:latest /app/otto serve -mock -local
```

#### With MQTT
```bash
podman run -p 8011:8011 otto:latest /app/otto serve -mock
```

#### Custom MQTT Broker
```bash
podman run -p 8011:8011 otto:latest /app/otto serve -mock -mqtt-broker your.broker.com
```

#### With Data Persistence
```bash
podman run \
  -p 8011:8011 \
  -v otto-data:/app/data \
  otto:latest /app/otto serve -mock
```

## Available Commands

View available commands:
```bash
podman run otto:latest /app/otto --help
```

Common commands:
- `serve` - Start the HTTP server
- `cli` - Interactive CLI mode
- `msg connect` - Connect to MQTT broker
- `msg pub` - Publish messages
- `msg sub` - Subscribe to topics
- `version` - Show version information

## Docker Compose (Optional)

Create a `docker-compose.yml` for easier management:

```yaml
version: '3.8'

services:
  otto:
    build: .
    ports:
      - "8011:8011"
    volumes:
      - otto-data:/app/data
    environment:
      - MQTT_BROKER=mqtt-broker
    command: /app/otto serve -mock -mqtt-broker mqtt-broker
    depends_on:
      - mqtt-broker
  
  mqtt-broker:
    image: eclipse-mosquitto:latest
    ports:
      - "1883:1883"
    volumes:
      - mqtt-data:/mosquitto/data

volumes:
  otto-data:
  mqtt-data:
```

Run with Docker Compose:
```bash
podman-compose up
```

## Environment Variables

Common environment variables you may want to set:
- `APP_DIR` - Web app root directory (default: embed)

## Ports

- **8011** - Web UI and REST API
- **1883** - MQTT broker (if running internal broker)

## Accessing the Application

Once running, access the web interface at:
```
http://localhost:8011
```

## Building for Different Architectures

### ARM (Raspberry Pi)
```bash
podman build --platform linux/arm/v7 -t otto:rpi-armv7 .
```

### ARM64 (Raspberry Pi 4+)
```bash
podman build --platform linux/arm64 -t otto:arm64 .
```

### AMD64 (Standard Linux/Windows)
```bash
podman build --platform linux/amd64 -t otto:amd64 .
```

## Advanced Usage

### Run in Detached Mode with Logs
```bash
podman run -d --name otto -p 8011:8011 otto:latest /app/otto serve -mock
podman logs -f otto
```

### Run Multiple Instances
```bash
# Station 1
podman run -d --name otto-1 -p 8011:8011 otto:latest /app/otto serve -mock

# Station 2
podman run -d --name otto-2 -p 8012:8011 otto:latest /app/otto serve -mock
```

### Interactive Shell
```bash
podman run -it otto:latest /bin/sh
```

## Troubleshooting

### Check Container Logs
```bash
podman logs <container-name>
```

### Verify Container is Running
```bash
podman ps
```

### Stop and Remove Container
```bash
podman stop otto
podman rm otto
```

### Prune Unused Images
```bash
podman image prune -a
```

## Security Notes

- The container runs as a non-root user (`otto`) for security
- Uses Alpine Linux for minimal attack surface
- Includes health checks for automatic restarts
- Mounts data directory for data persistence

## Building Notes

The Containerfile uses a multi-stage build to:
1. **Builder stage**: Compiles the Go application
2. **Runtime stage**: Minimal Alpine image with only runtime dependencies

This keeps the final image small (~20-30MB) while including all necessary functionality.
