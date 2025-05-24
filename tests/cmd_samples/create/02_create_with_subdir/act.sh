#!/bin/bash
set -e # Exit on first error

INIT_OUTPUT=$(./gydnc init)
# We don't echo INIT_OUTPUT here, as stdout assertion is for the file content

GYDNC_CONFIG_PATH=$(echo "$INIT_OUTPUT" | grep "export GYDNC_CONFIG" | head -n1 | cut -d'"' -f2)

if [ -z "$GYDNC_CONFIG_PATH" ]; then
  echo "Failed to extract GYDNC_CONFIG_PATH from init output:" >&2
  echo "$INIT_OUTPUT" >&2
  exit 1
fi

# Create the entity with a subdirectory in its alias, using the discovered config
./gydnc --config "$GYDNC_CONFIG_PATH" create category/my-sub-guidance --title ""

# Verify the content of the created file by catting it
echo "--- Content of .gydnc/category/my-sub-guidance.g6e: ---"
cat ".gydnc/category/my-sub-guidance.g6e"