#!/bin/sh

set -e

# gydnc installer script
# Usage: curl -sSL https://raw.githubusercontent.com/ofthemachine/gydnc/main/install.sh | sh

REPO_PATH="ofthemachine/gydnc" # Full path for API calls and URL construction
BINARY_NAME="gydnc"
CLONE_DIR_NAME=$(basename "$REPO_PATH") # Simple name for local directory, e.g., "gydnc"

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

echo "Determining latest gydnc version for $OS-$ARCH..."

# Get latest release tag from GitHub API
if command -v curl >/dev/null 2>&1; then
    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO_PATH/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^\"]+)\".*/\\1/')
elif command -v wget >/dev/null 2>&1; then
    LATEST_TAG=$(wget -qO- "https://api.github.com/repos/$REPO_PATH/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^\"]+)\".*/\\1/')
else
    echo "Error: curl or wget is required to determine the latest version."
    exit 1
fi

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not determine the latest release version."
    exit 1
fi

echo "Latest version is $LATEST_TAG"

# Construct asset names based on the fetched tag
# Assuming raw binaries are named like: gydnc-vX.Y.Z-os-arch
ASSET_BASENAME="$BINARY_NAME-$LATEST_TAG-$OS-$ARCH"
BINARY_FILENAME="$ASSET_BASENAME" # If they are raw binaries
# If assets are .tar.gz, this would be:
# ASSET_FILENAME="$ASSET_BASENAME.tar.gz"
# BINARY_IN_TAR="$BINARY_NAME" # Name of binary inside tarball

BINARY_URL="https://github.com/$REPO_PATH/releases/download/$LATEST_TAG/$BINARY_FILENAME"
SHA256_URL="https://github.com/$REPO_PATH/releases/download/$LATEST_TAG/$BINARY_FILENAME.sha256"

echo "Attempting to download pre-built binary from: $BINARY_URL"

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"
echo "Working in temporary directory: $TMP_DIR"

# Check if pre-built binary is available
if command -v curl >/dev/null 2>&1; then
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BINARY_URL")
elif command -v wget >/dev/null 2>&1; then
    # Wget spider doesn't follow redirects by default for http_code, so download and check
    wget --timeout=5 -q "$BINARY_URL" -O "$BINARY_FILENAME.tmpdownload"
    if [ $? -eq 0 ] && [ -f "$BINARY_FILENAME.tmpdownload" ]; then
        HTTP_CODE="200"
        rm "$BINARY_FILENAME.tmpdownload"
    else
        HTTP_CODE="404" # Or some other error
    fi
else
    echo "Error: curl or wget is required"
    rm -rf "$TMP_DIR"
    exit 1
fi

ACTUAL_BINARY_FILE_TO_INSTALL="$BINARY_FILENAME"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Pre-built binary ($BINARY_FILENAME) found, downloading..."

    # Download binary and checksum
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$BINARY_FILENAME" "$BINARY_URL"
        curl -L -o "$BINARY_FILENAME.sha256" "$SHA256_URL"
    else
        wget -O "$BINARY_FILENAME" "$BINARY_URL"
        wget -O "$BINARY_FILENAME.sha256" "$SHA256_URL"
    fi

    # Verify SHA256 checksum
    if [ -f "$BINARY_FILENAME.sha256" ]; then
        if command -v sha256sum >/dev/null 2>&1; then
            echo "🔍 Verifying checksum..."
            if sha256sum -c "$BINARY_FILENAME.sha256"; then
                echo "✅ Checksum verified"
            else
                echo "❌ Checksum verification failed for $BINARY_FILENAME"
                rm -rf "$TMP_DIR"
                exit 1
            fi
        elif command -v shasum >/dev/null 2>&1; then
            echo "🔍 Verifying checksum..."
            # shasum -c expects checksum and filename, ensure .sha256 file is formatted correctly or adapt
            # Assuming .sha256 contains "checksum  filename"
            # Create a temp checksum file in the format shasum expects if needed
            echo "$(cat $BINARY_FILENAME.sha256 | awk '{print $1}')  $BINARY_FILENAME" > $BINARY_FILENAME.shasum_check
            if shasum -a 256 -c "$BINARY_FILENAME.shasum_check"; then
                echo "✅ Checksum verified"
            else
                echo "❌ Checksum verification failed for $BINARY_FILENAME"
                rm -rf "$TMP_DIR"
                exit 1
            fi
            rm "$BINARY_FILENAME.shasum_check"
        else
            echo "⚠️  No SHA256 tool found, skipping checksum verification"
        fi
    else
        echo "⚠️  Checksum file $BINARY_FILENAME.sha256 not found. Skipping verification."
    fi


    # Check if binary exists
    if [ ! -f "$ACTUAL_BINARY_FILE_TO_INSTALL" ]; then
        echo "Error: Binary file $ACTUAL_BINARY_FILE_TO_INSTALL not found after download."
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Make executable
    chmod +x "$ACTUAL_BINARY_FILE_TO_INSTALL"

    # Install to system path
    INSTALL_DIR="/usr/local/bin"
    echo "Attempting to install $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$ACTUAL_BINARY_FILE_TO_INSTALL" "$INSTALL_DIR/$BINARY_NAME"
        echo "✅ $BINARY_NAME installed to $INSTALL_DIR/$BINARY_NAME"
    else
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$ACTUAL_BINARY_FILE_TO_INSTALL" "$INSTALL_DIR/$BINARY_NAME"
        echo "✅ $BINARY_NAME installed to $INSTALL_DIR/$BINARY_NAME"
    fi

else
    echo "⚠️  Pre-built binary not available (HTTP $HTTP_CODE for $BINARY_URL)"
    echo "📦 Building from source instead..."

    # Check for required tools
    if ! command -v git >/dev/null 2>&1; then
        echo "Error: git is required to build from source"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    if ! command -v go >/dev/null 2>&1; then
        echo "Error: Go (go) compiler is required to build from source"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    REPO_GIT_URL="https://github.com/$REPO_PATH.git"
    echo "Cloning repository $REPO_GIT_URL into $CLONE_DIR_NAME..."
    git clone --depth=1 "$REPO_GIT_URL" "$CLONE_DIR_NAME"

    if [ ! -d "$CLONE_DIR_NAME" ]; then
        echo "Error: Failed to clone repository into $CLONE_DIR_NAME."
        rm -rf "$TMP_DIR"
        exit 1
    fi

    cd "$CLONE_DIR_NAME"
    echo "Successfully changed directory to $(pwd)"

    if ! command -v make >/dev/null 2>&1; then
        # Try to build without make if not present, directly using go build
        echo "Warning: make is not installed. Attempting to build with 'go build'."
        echo "🔨 Building $BINARY_NAME..."
        CGO_ENABLED=0 go build -ldflags \"-s -w\" -o "$BINARY_NAME" ./cmd/gydnc
    else
        echo "🔨 Building $BINARY_NAME with make..."
        make build
    fi


    # Check if binary was built
    # The binary should be in the current directory ($CLONE_DIR_NAME) with the name $BINARY_NAME
    if [ ! -f "$BINARY_NAME" ]; then
        echo "Error: Failed to build $BINARY_NAME from source in $(pwd)."
        # cd .. before removing TMP_DIR, already in $CLONE_DIR_NAME, so cd out once more if needed by rm logic.
        cd .. # Back to $TMP_DIR
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Install
    INSTALL_DIR="/usr/local/bin"
    echo "Attempting to install $BINARY_NAME from $(pwd)/$BINARY_NAME to $INSTALL_DIR/$BINARY_NAME..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "$INSTALL_DIR/"
        echo "✅ $BINARY_NAME built and installed to $INSTALL_DIR/$BINARY_NAME"
    else
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
        echo "✅ $BINARY_NAME built and installed to $INSTALL_DIR/$BINARY_NAME"
    fi
    # cd out of cloned repo before TMP_DIR removal
    cd .. # Back to $TMP_DIR
fi

# Cleanup
echo "Cleaning up temporary directory: $TMP_DIR"
rm -rf "$TMP_DIR"

# Verify installation
if command -v $BINARY_NAME >/dev/null 2>&1; then
    echo "🎉 Installation successful!"
    echo "Run '$BINARY_NAME --help' to get started."
else
    echo "⚠️  Installation completed, but $BINARY_NAME is not in your PATH."
    echo "You may need to restart your shell or add $INSTALL_DIR to your PATH."
fi