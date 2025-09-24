#!/bin/bash
set -e

# linctl DevContainer Feature Installation Script
# This script installs linctl CLI for Linear project management

echo "🔧 Installing linctl..."

# Parse options from environment variables
VERSION="${VERSION:-latest}"
INSTALL_METHOD="${INSTALLMETHOD:-release}"

# Detect architecture
ARCH=$(dpkg --print-architecture 2>/dev/null || echo "amd64")
case $ARCH in
    "amd64") ARCH="amd64" ;;
    "arm64") ARCH="arm64" ;;
    "armhf") ARCH="arm64" ;;  # Fallback for ARM
    *)
        echo "⚠️  Unsupported architecture: $ARCH, falling back to source build"
        INSTALL_METHOD="source"
        ;;
esac

# Function to install from GitHub releases
install_from_release() {
    echo "📦 Installing linctl from GitHub releases..."

    # Get latest release info if version is 'latest'
    if [ "$VERSION" = "latest" ]; then
        echo "🔍 Fetching latest release information..."
        RELEASE_INFO=$(curl -s https://api.github.com/repos/nicholls-inc/linctl/releases/latest)
        VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)

        if [ -z "$VERSION" ] || [ "$VERSION" = "null" ]; then
            echo "❌ Failed to fetch latest release version, falling back to source build"
            install_from_source
            return
        fi

        echo "✅ Latest version: $VERSION"
    fi

    # Construct download URL
    BINARY_NAME="linctl-linux-${ARCH}"
    DOWNLOAD_URL="https://github.com/nicholls-inc/linctl/releases/download/${VERSION}/${BINARY_NAME}"
    CHECKSUMS_URL="https://github.com/nicholls-inc/linctl/releases/download/${VERSION}/checksums.txt"

    echo "📥 Downloading linctl binary..."
    echo "   URL: $DOWNLOAD_URL"

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # Download binary
    if ! curl -L -f -o "$BINARY_NAME" "$DOWNLOAD_URL"; then
        echo "❌ Failed to download binary, falling back to source build"
        rm -rf "$TMP_DIR"
        install_from_source
        return
    fi

    # Download and verify checksums if available
    if curl -L -f -o checksums.txt "$CHECKSUMS_URL" 2>/dev/null; then
        echo "🔐 Verifying checksums..."
        if sha256sum -c --ignore-missing checksums.txt; then
            echo "✅ Checksum verification passed"
        else
            echo "❌ Checksum verification failed, falling back to source build"
            rm -rf "$TMP_DIR"
            install_from_source
            return
        fi
    else
        echo "⚠️  Checksums not available, skipping verification"
    fi

    # Install binary
    chmod +x "$BINARY_NAME"
    mv "$BINARY_NAME" /usr/local/bin/linctl

    # Cleanup
    cd /
    rm -rf "$TMP_DIR"

    echo "✅ linctl installed successfully from release"
}

# Function to install from source
install_from_source() {
    echo "🔨 Installing linctl from source..."

    # Check if Go is available
    if ! command -v go >/dev/null 2>&1; then
        echo "❌ Go is not installed. Installing Go..."

        # Install Go if not present
        GO_VERSION="1.23.11"
        GO_ARCH="amd64"
        if [ "$ARCH" = "arm64" ]; then
            GO_ARCH="arm64"
        fi

        GO_TARBALL="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
        curl -L -o "/tmp/${GO_TARBALL}" "https://golang.org/dl/${GO_TARBALL}"
        tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
        export PATH="/usr/local/go/bin:$PATH"
        rm "/tmp/${GO_TARBALL}"
    fi

    # Create temporary directory for source build
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # Clone the repository
    echo "📥 Cloning linctl repository..."
    git clone https://github.com/nicholls-inc/linctl.git .

    # Checkout specific version if not latest and not empty
    if [ "$VERSION" != "latest" ] && [ -n "$VERSION" ]; then
        echo "🔄 Checking out version $VERSION..."
        git checkout "$VERSION" || {
            echo "❌ Failed to checkout version $VERSION, using main branch"
        }
    fi

    # Build the binary
    echo "🔨 Building linctl..."
    make deps
    make build

    # Install binary
    chmod +x linctl
    mv linctl /usr/local/bin/linctl

    # Cleanup
    cd /
    rm -rf "$TMP_DIR"

    echo "✅ linctl built and installed successfully from source"
}

# Main installation logic
case "$INSTALL_METHOD" in
    "release")
        install_from_release
        ;;
    "source")
        install_from_source
        ;;
    *)
        echo "❌ Invalid install method: $INSTALL_METHOD"
        echo "   Valid options: release, source"
        exit 1
        ;;
esac

# Verify installation
if command -v linctl >/dev/null 2>&1; then
    echo "🎉 linctl installation completed successfully!"
    echo "📋 Version information:"
    linctl --version
    echo ""
    echo "💡 Usage: linctl --help"
    echo "🔗 Documentation: https://github.com/nicholls-inc/linctl"
else
    echo "❌ linctl installation failed - binary not found in PATH"
    exit 1
fi
