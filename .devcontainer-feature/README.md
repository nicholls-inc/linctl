# Linear CLI (linctl) DevContainer Feature

This DevContainer Feature installs [linctl](https://github.com/nicholls-inc/linctl), a command-line interface for Linear project management.

## Usage

Add this feature to your `devcontainer.json`:

```json
{
    "features": {
        "ghcr.io/nicholls-inc/linctl/linctl:1": {}
    }
}
```

## Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `version` | string | `latest` | Version of linctl to install. Use 'latest' for the most recent release or specify a version tag (e.g., 'v1.0.0') |
| `installMethod` | string | `release` | Installation method: 'release' downloads pre-built binaries, 'source' builds from source code |

## Examples

### Install Latest Version (Default)

```json
{
    "features": {
        "ghcr.io/nicholls-inc/linctl/linctl:1": {}
    }
}
```

### Install Specific Version

```json
{
    "features": {
        "ghcr.io/nicholls-inc/linctl/linctl:1": {
            "version": "v1.0.0"
        }
    }
}
```

### Build from Source

```json
{
    "features": {
        "ghcr.io/nicholls-inc/linctl/linctl:1": {
            "installMethod": "source"
        }
    }
}
```

## Installation Methods

### Release (Default)
- Downloads pre-built binaries from GitHub releases
- Faster installation
- Includes checksum verification for security
- Supports linux/amd64 and linux/arm64 architectures
- Falls back to source build if release is unavailable

### Source
- Builds linctl from source code
- Requires Go compiler (automatically installed if missing)
- Always uses the latest code from the repository
- Useful for development or when pre-built binaries are unavailable

## Authentication

After installation, you'll need to authenticate with Linear:

```bash
# OAuth authentication (recommended)
linctl auth login --oauth

# API key authentication
linctl auth login --api-key YOUR_API_KEY
```

## Supported Architectures

- linux/amd64
- linux/arm64

## Requirements

- Linux-based container
- Internet access for downloading binaries or source code
- For source builds: Go 1.23+ (automatically installed if missing)

## Verification

After installation, verify linctl is working:

```bash
linctl --version
linctl --help
```

## Related

- [linctl GitHub Repository](https://github.com/nicholls-inc/linctl)
- [Linear API Documentation](https://developers.linear.app/)
- [DevContainer Features Specification](https://containers.dev/implementors/features/)
