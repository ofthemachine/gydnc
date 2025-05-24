#!/bin/sh

set -e

# gydnc installer script
# Usage: curl -sSL https://raw.githubusercontent.com/ofthemachine/gydnc/main/install.sh | sh

REPO="ofthemachine/gydnc"
BINARY="gydnc"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture names
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Validate platform
case $OS in
    linux|darwin)
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

echo "Installing gydnc for $OS-$ARCH..."

# Try to download pre-built binary first
BINARY_URL="https://github.com/$REPO/releases/latest/download/$BINARY-$OS-$ARCH"
SHA256_URL="https://github.com/$REPO/releases/latest/download/$BINARY-$OS-$ARCH.sha256"
echo "Attempting to download pre-built binary from: $BINARY_URL"

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Check if pre-built binary is available
if command -v curl >/dev/null 2>&1; then
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BINARY_URL")
elif command -v wget >/dev/null 2>&1; then
    HTTP_CODE=$(wget --spider --server-response "$BINARY_URL" 2>&1 | grep "HTTP/" | tail -1 | awk '{print $2}')
else
    echo "Error: curl or wget is required"
    exit 1
fi

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Pre-built binary found, downloading..."

    # Download binary and checksum
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$BINARY" "$BINARY_URL"
        curl -L -o "$BINARY.sha256" "$SHA256_URL"
    else
        wget -O "$BINARY" "$BINARY_URL"
        wget -O "$BINARY.sha256" "$SHA256_URL"
    fi

    # Verify SHA256 checksum
    if command -v sha256sum >/dev/null 2>&1; then
        echo "🔍 Verifying checksum..."
        if sha256sum -c "$BINARY.sha256"; then
            echo "✅ Checksum verified"
        else
            echo "❌ Checksum verification failed"
            exit 1
        fi
    elif command -v shasum >/dev/null 2>&1; then
        echo "🔍 Verifying checksum..."
        if shasum -a 256 -c "$BINARY.sha256"; then
            echo "✅ Checksum verified"
        else
            echo "❌ Checksum verification failed"
            exit 1
        fi
    else
        echo "⚠️  No SHA256 tool found, skipping checksum verification"
    fi

    # Check if binary exists
    if [ ! -f "$BINARY" ]; then
        echo "Error: Binary not found"
        exit 1
    fi

    # Make executable
    chmod +x "$BINARY"

    # Install to system path
    INSTALL_DIR="/usr/local/bin"
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY" "$INSTALL_DIR/"
        echo "✅ gydnc installed to $INSTALL_DIR/$BINARY"
    else
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$BINARY" "$INSTALL_DIR/"
        echo "✅ gydnc installed to $INSTALL_DIR/$BINARY"
    fi

else
    echo "⚠️  Pre-built binary not available (HTTP $HTTP_CODE)"
    echo "📦 Building from source instead..."

    # Check for required tools
    if ! command -v git >/dev/null 2>&1; then
        echo "Error: git is required to build from source"
        exit 1
    fi

    if ! command -v make >/dev/null 2>&1; then
        echo "Error: make is required to build from source"
        exit 1
    fi

    # Clone and build
    git clone "https://github.com/$REPO.git"
    cd "$REPO"
    echo "🔨 Building gydnc..."
    make build

    # Install
    INSTALL_DIR="/usr/local/bin"
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY" "$INSTALL_DIR/"
        echo "✅ gydnc built and installed to $INSTALL_DIR/$BINARY"
    else
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$BINARY" "$INSTALL_DIR/"
        echo "✅ gydnc built and installed to $INSTALL_DIR/$BINARY"
    fi
fi

# Cleanup
cd /
rm -rf "$TMP_DIR"

# Verify installation
if command -v gydnc >/dev/null 2>&1; then
    echo "🎉 Installation successful!"
    echo "Run 'gydnc --help' to get started."
else
    echo "⚠️  Installation completed, but gydnc is not in your PATH."
    echo "You may need to restart your shell or add $INSTALL_DIR to your PATH."
fi