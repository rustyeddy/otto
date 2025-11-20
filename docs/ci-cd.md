# CI/CD Pipeline Documentation

## Overview

The Otto project uses GitHub Actions for continuous integration and continuous deployment. The pipeline automatically builds, tests, and creates release packages for multiple platforms.

## Workflows

### CI Pipeline (`ci.yml`)

Triggers on:
- Push to `main` branch
- Pull requests to `main` branch

Steps:
1. **Format Check** - Ensures all code is properly formatted with `go fmt`
2. **Tests** - Runs the complete test suite with coverage reporting
3. **Multi-platform Builds** - Creates binaries for all supported platforms

Supported Platforms:
- Linux x86_64 (`otto-linux-amd64`)
- Linux ARM v7 (Raspberry Pi) (`otto-linux-arm7`)
- Linux ARM64 (Raspberry Pi 4+) (`otto-linux-arm64`)
- macOS x86_64 (`otto-darwin-amd64`)
- macOS ARM64 (Apple Silicon) (`otto-darwin-arm64`)
- Windows x86_64 (`otto-windows-amd64.exe`)

Artifacts are retained for 30 days and can be downloaded from the Actions tab.

### Release Pipeline (`release.yml`)

Triggers on:
- Git tags matching pattern `v*` (e.g., `v0.1.0`, `v1.0.0`)

Steps:
1. Builds binaries for all supported platforms
2. Creates SHA256 checksums for each binary
3. Creates a GitHub release with all binaries and checksums attached

## Creating a Release

To create a new release:

```bash
# Tag the release
git tag -a v0.2.0 -m "Release version 0.2.0"

# Push the tag
git push origin v0.2.0
```

The release workflow will automatically:
- Build binaries for all platforms
- Create checksums
- Create a GitHub release with downloadable assets

## Local Development

### Prerequisites
- Go 1.23 or later

### Building Locally

```bash
# Build for current platform
make build

# Run tests
make test

# Format code
make fmt

# Run complete CI checks locally
make ci
```

### Cross-platform Building

Build for specific platforms:

```bash
# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o otto-linux-amd64 ./cmd/otto

# Raspberry Pi (ARM v7)
GOOS=linux GOARCH=arm GOARM=7 go build -o otto-linux-arm7 ./cmd/otto

# Raspberry Pi 4+ (ARM64)
GOOS=linux GOARCH=arm64 go build -o otto-linux-arm64 ./cmd/otto

# macOS x86_64
GOOS=darwin GOARCH=amd64 go build -o otto-darwin-amd64 ./cmd/otto

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o otto-darwin-arm64 ./cmd/otto

# Windows
GOOS=windows GOARCH=amd64 go build -o otto-windows-amd64.exe ./cmd/otto
```

## Downloading Releases

### From GitHub Releases

Visit the [Releases page](https://github.com/rustyeddy/otto/releases) to download the latest version for your platform.

### From CI Artifacts

For development builds:
1. Go to the Actions tab
2. Select a workflow run
3. Download the artifact for your platform

## Verifying Downloads

Each release includes SHA256 checksums. To verify a download:

```bash
# Linux/macOS
sha256sum -c otto-linux-amd64.sha256

# Or manually compare
sha256sum otto-linux-amd64
cat otto-linux-amd64.sha256
```

## Platform-specific Installation

### Linux (x86_64)
```bash
wget https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-linux-amd64
chmod +x otto-linux-amd64
sudo mv otto-linux-amd64 /usr/local/bin/otto
```

### Raspberry Pi (ARM v7)
```bash
wget https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-linux-arm7
chmod +x otto-linux-arm7
sudo mv otto-linux-arm7 /usr/local/bin/otto
```

### Raspberry Pi 4+ (ARM64)
```bash
wget https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-linux-arm64
chmod +x otto-linux-arm64
sudo mv otto-linux-arm64 /usr/local/bin/otto
```

### macOS
```bash
# x86_64
curl -LO https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-darwin-amd64
chmod +x otto-darwin-amd64
mv otto-darwin-amd64 /usr/local/bin/otto

# Apple Silicon
curl -LO https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-darwin-arm64
chmod +x otto-darwin-arm64
mv otto-darwin-arm64 /usr/local/bin/otto
```

### Windows
Download `otto-windows-amd64.exe` from the releases page and add it to your PATH.

## Troubleshooting

### Format Check Fails

If the CI format check fails:
```bash
make fmt
git add .
git commit -m "Format code"
git push
```

### Build Fails

Check the build logs in the Actions tab for specific errors. Common issues:
- Missing dependencies: Run `go mod tidy`
- Import errors: Ensure all imports are correct
- Go version: Ensure you're using Go 1.23 or later

### Tests Fail

Run tests locally to debug:
```bash
make test          # Run all tests
make verbose       # Run with verbose output
make coverage      # See coverage report
```
