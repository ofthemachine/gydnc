#!/bin/bash
set -e # Exit on first error

# Perform init
INIT_OUTPUT=$(./gydnc init)
echo "$INIT_OUTPUT" # Echo the init output to be captured by stdout assertion

# Extract config path and list
GYDNC_CONFIG_PATH=$(echo "$INIT_OUTPUT" | grep "export GYDNC_CONFIG" | head -n1 | cut -d'"' -f2)
if [ -n "$GYDNC_CONFIG_PATH" ]; then
  ./gydnc --config "$GYDNC_CONFIG_PATH" list > /dev/null # Run list, but discard its stdout for this test's stdout assertion
else
  echo "Failed to extract GYDNC_CONFIG_PATH" >&2
  exit 1
fi