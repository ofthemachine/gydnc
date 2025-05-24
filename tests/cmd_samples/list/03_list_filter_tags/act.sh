#!/bin/bash
set -euo pipefail

# Setup: Use a shared config and create some entities with tags
TEST_DIR=$(pwd)
CONFIG_CONTENT="default_backend: primary\nstorage_backends:\n  primary:\n    type: localfs\n    localfs:\n      path: $TEST_DIR/test_data\n"
mkdir -p .gydnc test_data
echo -e "$CONFIG_CONTENT" > .gydnc/config.yml
export GYDNC_CONFIG="$TEST_DIR/.gydnc/config.yml"

./gydnc create entityA --title "Entity A" --tags "urgent,feat" > /dev/null
./gydnc create entityB --title "Entity B" --tags "urgent,bug" > /dev/null
./gydnc create entityC --title "Entity C" --tags "feat,test" > /dev/null

# List with tag filter
./gydnc list --filter-tags "urgent"

echo "---Filtering for feat AND NOT bug---"
./gydnc list --filter-tags "feat -bug"
