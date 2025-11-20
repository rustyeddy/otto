# CI/CD Quick Start Guide

## What Was Added

The Otto project now has a complete CI/CD pipeline that automatically:
1. ✅ Tests your code on every push and PR
2. ✅ Builds binaries for 6 platforms (Linux, macOS, Windows, Raspberry Pi)
3. ✅ Creates releases with downloadable binaries

## For Users: Getting Otto Binaries

### Option 1: Download from Releases (Recommended)
After the first release is created, visit:
https://github.com/rustyeddy/otto/releases

Download the binary for your platform:
- `otto-linux-amd64` - Linux x86_64
- `otto-linux-arm7` - Raspberry Pi (older models)
- `otto-linux-arm64` - Raspberry Pi 4+
- `otto-darwin-amd64` - macOS Intel
- `otto-darwin-arm64` - macOS Apple Silicon
- `otto-windows-amd64.exe` - Windows

### Option 2: Download from CI Artifacts
For development builds:
1. Go to the [Actions tab](https://github.com/rustyeddy/otto/actions)
2. Click on a successful workflow run
3. Scroll down to "Artifacts"
4. Download the binary for your platform

### Installation Example (Linux/macOS)
```bash
# Download (replace URL with actual release)
curl -LO https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-linux-amd64

# Make executable
chmod +x otto-linux-amd64

# Move to PATH (optional)
sudo mv otto-linux-amd64 /usr/local/bin/otto

# Verify
otto version
```

## For Developers: Building Locally

### Quick Build
```bash
# Clone the repository
git clone https://github.com/rustyeddy/otto
cd otto

# Build for your platform
make build

# Run it
./otto version
```

### Available Make Commands
```bash
make build      # Build otto binary
make test       # Run all tests
make fmt        # Format all code
make vet        # Run go vet (linter)
make clean      # Remove build artifacts
make ci         # Run complete CI checks
```

### Cross-Platform Builds
```bash
# Raspberry Pi
GOOS=linux GOARCH=arm GOARM=7 go build -o otto-rpi ./cmd/otto

# Windows
GOOS=windows GOARCH=amd64 go build -o otto.exe ./cmd/otto

# macOS
GOOS=darwin GOARCH=arm64 go build -o otto-mac ./cmd/otto
```

## For Maintainers: Creating Releases

### Step-by-Step Release Process
```bash
# 1. Ensure everything is committed and pushed
git status
git push

# 2. Create a version tag (use semantic versioning)
git tag -a v0.2.0 -m "Release version 0.2.0 - Add awesome feature"

# 3. Push the tag
git push origin v0.2.0

# 4. Wait 2-3 minutes, then check:
# https://github.com/rustyeddy/otto/releases
```

That's it! The release workflow automatically:
- Builds binaries for all 6 platforms
- Generates SHA256 checksums
- Creates a GitHub release
- Attaches all binaries as downloadable assets

### Pre-release Versions
For alpha/beta/rc releases:
```bash
git tag -a v0.2.0-beta.1 -m "Beta release"
git push origin v0.2.0-beta.1
```

These will be marked as "pre-release" on GitHub.

## Continuous Integration (CI)

Every push and pull request automatically:
1. **Format Check** - Ensures code is properly formatted
2. **Tests** - Runs the complete test suite
3. **Builds** - Creates binaries for all platforms

If any step fails, you'll see a ❌ on GitHub and in the PR.

### Fixing CI Failures

**Format Check Failed:**
```bash
make fmt
git add .
git commit -m "Format code"
git push
```

**Tests Failed:**
```bash
make test          # Run locally to debug
make verbose       # See detailed test output
```

**Build Failed:**
```bash
make build         # Try building locally
go mod tidy        # Update dependencies if needed
```

## What Changed in the Repository

### New Files
- `cmd/otto/main.go` - CLI application entry point
- `.github/workflows/ci.yml` - CI pipeline
- `.github/workflows/release.yml` - Release automation
- `docs/ci-cd.md` - Comprehensive CI/CD documentation
- `docs/RELEASE.md` - Release process guide
- `docs/QUICKSTART-CICD.md` - This file!

### Updated Files
- `Makefile` - Added build, fmt, clean, ci targets
- `README.md` - Added download and build instructions
- `.gitignore` - Exclude binary files

### Removed Files
- `.github/workflows/build-otto.yml` - Replaced by ci.yml
- `.github/workflows/run-tests.yml` - Replaced by ci.yml

## Platform Support

| Platform | Architecture | Example Use Case |
|----------|-------------|------------------|
| Linux x86_64 | amd64 | Servers, desktops |
| Linux ARM v7 | armv7 | Raspberry Pi 2, 3 |
| Linux ARM64 | arm64 | Raspberry Pi 4+ |
| macOS Intel | amd64 | Intel Macs |
| macOS ARM | arm64 | Apple Silicon Macs |
| Windows | amd64 | Windows desktops |

## Need Help?

- **Documentation**: See `docs/ci-cd.md` for detailed information
- **Release Guide**: See `docs/RELEASE.md` for release procedures
- **Issues**: Open an issue on GitHub
- **CI Status**: Check the Actions tab on GitHub

## Next Steps

1. **Create First Release**: Tag v0.2.0 to test the release workflow
2. **Download and Test**: Verify binaries work on different platforms
3. **Update Documentation**: Add platform-specific setup guides
4. **Celebrate**: You now have automated CI/CD! 🎉
