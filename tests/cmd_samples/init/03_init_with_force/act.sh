#!/bin/bash
set -euo pipefail # Exit on error, fail on pipe errors

# First initialization - should succeed normally
echo "=== First initialization ==="
./gydnc init
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"

# Create a dummy entity to confirm the first init worked
echo "# Test Entity" | ./gydnc create test-entity --title "Test Entity" > /dev/null

# Confirm the initial setup worked
./gydnc list

# Now attempt to initialize again - this should fail without --force
echo ""
echo "=== Second initialization without --force (should fail) ==="
if ./gydnc init 2>/dev/null; then
  echo "ERROR: Second init without --force succeeded unexpectedly!"
  exit 1
else
  echo "Second init without --force failed as expected"
fi

# Finally, initialize again with --force - this should succeed
echo ""
echo "=== Third initialization with --force (should succeed) ==="
./gydnc init --force

# Verify we can still list entities
echo ""
echo "=== Verify list still works after force init ==="
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"
./gydnc list