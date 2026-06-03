#!/bin/bash
# install.sh - Install ostrakon from GitHub releases
# Usage: curl -sSL https://raw.githubusercontent.com/PapaDanielVi/ostrakon/main/install.sh | bash

set -euo pipefail

# Configuration
REPO="PapaDanielVi/ostrakon"
BINARY_NAME="ostrakon"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect architecture
ARCH=$(uname -m)

# Map architecture to release naming
case "$ARCH" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# Map architecture to release naming
case "$ARCH" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# Detect latest version
echo "Detecting latest version..."
LATEST_VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo "Failed to detect latest version" >&2
    exit 1
fi

echo "Latest version: $LATEST_VERSION"

# Construct download URL
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}_Linux_${ARCH}.tar.gz"

# Check if binary already exists
if command -v "$BINARY_NAME" &> /dev/null; then
    CURRENT_VERSION=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
    echo "Current installed version: $CURRENT_VERSION"
    echo "Updating to latest version..."
else
    echo "Installing ostrakon..."
fi

# Create temporary directory for download
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
echo "Downloading from: $DOWNLOAD_URL"
curl -sSL "$DOWNLOAD_URL" -o "$TMP_DIR/ostrakon.tar.gz"

# Extract binary
tar -xzf "$TMP_DIR/ostrakon.tar.gz" -C "$TMP_DIR"

# Install binary
if [ -w "$INSTALL_DIR" ] || [ "$INSTALL_DIR" = "/usr/local/bin" ] && sudo -n true 2>/dev/null; then
    INSTALL_CMD="install"
else
    INSTALL_CMD="sudo install"
fi

$INSTALL_CMD -m 0755 "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

# Verify installation
echo "Verifying installation..."
"$INSTALL_DIR/$BINARY_NAME" --version

echo "Installation complete! ostrakon is installed to $INSTALL_DIR/$BINARY_NAME"
echo "Run 'ostrakon --help' to get started."