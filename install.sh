#!/bin/sh
set -e

# Guardian installer
# Usage: curl -fsSL https://raw.githubusercontent.com/cglabs-ai/guardian/main/install.sh | sh

REPO="cglabs-ai/guardian"
VERSION="v1.0.0"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY="guardian-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}.tar.gz"

# Install location
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -d "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
else
    mkdir -p "$HOME/.local/bin"
    INSTALL_DIR="$HOME/.local/bin"
fi

echo "Installing Guardian..."
echo "  OS: $OS"
echo "  Arch: $ARCH"
echo "  Installing to: $INSTALL_DIR"

# Download and extract
TMP=$(mktemp -d)
curl -fsSL "$URL" | tar -xz -C "$TMP"
mv "$TMP/$BINARY" "$INSTALL_DIR/guardian"
chmod +x "$INSTALL_DIR/guardian"
rm -rf "$TMP"

echo ""
echo "Guardian installed successfully!"
echo ""

# Check if in PATH
if command -v guardian >/dev/null 2>&1; then
    echo "Run 'guardian' to get started."
else
    echo "Add $INSTALL_DIR to your PATH:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
