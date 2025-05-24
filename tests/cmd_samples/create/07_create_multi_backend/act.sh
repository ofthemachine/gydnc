#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.
# Setup: Create a temporary gydnc config and two backend directories
TEST_DIR=$(pwd)
CONFIG_CONTENT="default_backend: backend1\nstorage_backends:\n  backend1:\n    type: localfs\n    localfs:\n      path: $TEST_DIR/backend1_data\n  backend2:\n    type: localfs\n    localfs:\n      path: $TEST_DIR/backend2_data\n"
mkdir -p .gydnc backend1_data backend2_data
echo -e "$CONFIG_CONTENT" > .gydnc/config.yml
export GYDNC_CONFIG="$TEST_DIR/.gydnc/config.yml"
export GYDNC_LOG_LEVEL="debug" # Enable debug logging for more test output

# Create guidance entity in backend1
cat << 'EOF' | ./gydnc create test-entity --title "Test Entity BE1" --backend "backend1"
This is test content that should appear in backend1.
EOF

# Create guidance entity in backend2
cat << 'EOF' | ./gydnc create test-entity --title "Test Entity BE2" --backend "backend2"
This is test content that should appear in backend2.
EOF

# List contents of both backends to verify
echo "=== Listing backend1 contents ==="
./gydnc list --backend backend1

# List contents of backend2 to verify (should only contain BE2 version)
echo "=== Listing backend2 contents ==="
./gydnc list --backend backend2

# List merged contents (no --backend flag) to verify de-duplication
echo "=== Listing merged contents (no backend flag) ==="
./gydnc list