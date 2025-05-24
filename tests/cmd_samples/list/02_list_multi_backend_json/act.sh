#!/bin/bash
set -euo pipefail

# Setup as in create/07_create_multi_backend
TEST_DIR=$(pwd)
CONFIG_CONTENT="default_backend: backend1\nstorage_backends:\n  backend1:\n    type: localfs\n    localfs:\n      path: $TEST_DIR/backend1_data\n  backend2:\n    type: localfs\n    localfs:\n      path: $TEST_DIR/backend2_data\n"
mkdir -p .gydnc backend1_data backend2_data
echo -e "$CONFIG_CONTENT" > .gydnc/config.yml
export GYDNC_CONFIG="$TEST_DIR/.gydnc/config.yml"
# export GYDNC_LOG_LEVEL="debug"

# Create entities
./gydnc create entity1 --title "Entity 1 in BE1" --backend backend1 > /dev/null
./gydnc create entity2 --title "Entity 2 in BE1" --backend backend1 > /dev/null
./gydnc create entity1 --title "Entity 1 in BE2" --backend backend2 > /dev/null # Duplicate alias
./gydnc create entity3 --title "Entity 3 in BE2" --backend backend2 > /dev/null

# List all (merged, default is JSON)
./gydnc list