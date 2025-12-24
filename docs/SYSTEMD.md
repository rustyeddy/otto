# Systemd Service Configuration

This document describes how to install and configure OttO as a systemd service.

## Installation

### 1. Build the Binary

```bash
make build
```

### 2. Install the Binary

Copy the binary to the installation location:

```bash
sudo mkdir -p /opt/otto/bin
sudo cp otto /opt/otto/bin/otto
sudo chmod +x /opt/otto/bin/otto
```

Or use the Makefile:

```bash
make install
```

### 3. Create System User

Create a dedicated user for running the otto service:

```bash
sudo useradd -r -s /bin/false -d /opt/otto otto
```

### 4. Create Working Directory

```bash
sudo mkdir -p /opt/otto/data
sudo chown -R otto:otto /opt/otto
```

### 5. Install Systemd Service

Copy the service file to the systemd directory:

```bash
sudo cp otto.service /etc/systemd/system/
sudo systemctl daemon-reload
```

Or use the Makefile:

```bash
make install-service
```

To install and enable the service in one step:

```bash
make install-service
make enable-service
```

## Quick Install with Makefile

The repository includes Makefile targets for easy installation:

```bash
# Build and install otto binary
make install

# Install systemd service
make install-service

# Enable and start the service
make enable-service

# Check service status
make service-status

# Follow service logs
make service-logs

# Uninstall everything
make uninstall
```

## Configuration

### Environment Variables

Edit the service file to customize environment variables:

```bash
sudo systemctl edit otto.service
```

Add your configuration:

```ini
[Service]
Environment="OTTO_LOG_LEVEL=debug"
Environment="OTTO_CONFIG=/etc/otto/config.yaml"
```

### Data Directory

By default, the service uses `/opt/otto/data` for data storage. To change this:

1. Edit the service file
2. Update `WorkingDirectory` and `ReadWritePaths`
3. Reload systemd: `sudo systemctl daemon-reload`

## Usage

### Enable Service (Start on Boot)

```bash
sudo systemctl enable otto.service
```

### Start Service

```bash
sudo systemctl start otto.service
```

### Stop Service

```bash
sudo systemctl stop otto.service
```

### Restart Service

```bash
sudo systemctl restart otto.service
```

### Check Service Status

```bash
sudo systemctl status otto.service
```

### View Logs

```bash
# View all logs
sudo journalctl -u otto.service

# Follow logs in real-time
sudo journalctl -u otto.service -f

# View logs from last boot
sudo journalctl -u otto.service -b

# View last 100 lines
sudo journalctl -u otto.service -n 100
```

## Security Features

The service file includes several security hardening features:

- **NoNewPrivileges**: Prevents the process from gaining new privileges
- **PrivateTmp**: Uses a private /tmp directory
- **ProtectSystem=strict**: Makes most of the filesystem read-only
- **ProtectHome**: Denies access to /home directories
- **ReadWritePaths**: Only allows writes to the data directory

## Troubleshooting

### Service Fails to Start

Check the logs for errors:

```bash
sudo journalctl -u otto.service -n 50 --no-pager
```

### Permission Issues

Ensure the otto user has proper permissions:

```bash
sudo chown -R otto:otto /opt/otto
sudo chmod -R 750 /opt/otto
```

### Binary Not Found

Verify the binary location matches the ExecStart path:

```bash
ls -l /opt/otto/bin/otto
```

## Customization

### Running as a Different User

If you want to run otto as a different user, edit the service file:

```bash
sudo systemctl edit --full otto.service
```

Change the `User` and `Group` directives, then reload:

```bash
sudo systemctl daemon-reload
sudo systemctl restart otto.service
```

### Custom Configuration File

To use a custom configuration file location:

```bash
sudo systemctl edit otto.service
```

Add:

```ini
[Service]
Environment="OTTO_CONFIG=/etc/otto/config.yaml"
```

Reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart otto.service
```

## Uninstallation

To remove the service:

```bash
sudo systemctl stop otto.service
sudo systemctl disable otto.service
sudo rm /etc/systemd/system/otto.service
sudo systemctl daemon-reload
```

To also remove the user and data:

```bash
sudo userdel otto
sudo rm -rf /opt/otto
```

Or simply use:

```bash
make uninstall
```
