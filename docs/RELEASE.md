# Release Process

## Quick Start

To create a new release of Otto with versioned binaries:

```bash
# 1. Update version in cmd/cmd_version.go if needed
# 2. Commit your changes
git add .
git commit -m "Prepare for release v0.2.0"
git push

# 3. Create and push a tag
git tag -a v0.2.0 -m "Release version 0.2.0"
git push origin v0.2.0
```

The release workflow will automatically:
- Build binaries for all 6 platforms
- Generate SHA256 checksums
- Create a GitHub release with all assets

## Version Numbering

Otto follows [Semantic Versioning](https://semver.org/):
- **Major**: Breaking changes (v1.0.0 → v2.0.0)
- **Minor**: New features, backwards compatible (v1.0.0 → v1.1.0)
- **Patch**: Bug fixes (v1.0.0 → v1.0.1)

## Pre-release Versions

For alpha, beta, or release candidate versions:

```bash
git tag -a v0.2.0-alpha.1 -m "Alpha release"
git push origin v0.2.0-alpha.1
```

These will be marked as pre-releases on GitHub.

## Release Checklist

Before creating a release:

- [ ] All tests pass: `make test`
- [ ] Code is formatted: `make fmt`
- [ ] Update version in `cmd/cmd_version.go`
- [ ] Update CHANGELOG.md (if exists)
- [ ] Update README.md if needed
- [ ] Create PR and get it merged
- [ ] Create and push tag from main branch
- [ ] Verify release artifacts on GitHub

## Platform Support

Each release includes binaries for:
- Linux x86_64 (servers, desktops)
- Linux ARM v7 (Raspberry Pi 2, 3)
- Linux ARM64 (Raspberry Pi 4+)
- macOS x86_64 (Intel Macs)
- macOS ARM64 (Apple Silicon Macs)
- Windows x86_64

## Verifying Releases

Each binary includes a SHA256 checksum file. To verify:

```bash
# Download binary and checksum
wget https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-linux-amd64
wget https://github.com/rustyeddy/otto/releases/download/v0.2.0/otto-linux-amd64.sha256

# Verify
sha256sum -c otto-linux-amd64.sha256
```

## Troubleshooting

### Release Failed

Check the Actions tab for error details. Common issues:
- Permission errors: Ensure GITHUB_TOKEN has proper permissions
- Build failures: Test locally first with cross-platform builds
- Tag conflicts: Delete and recreate the tag if needed

### Deleting a Failed Release

```bash
# Delete local tag
git tag -d v0.2.0

# Delete remote tag
git push --delete origin v0.2.0

# Delete the GitHub release manually from the web interface
# Then recreate with the same tag
```
