#!/bin/bash
set -euo pipefail

# Arrange: create a guidance entity
./gydnc init > /dev/null 2>&1 # Suppress init output
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"
./gydnc create simple-delete-test --title "Delete Me" --body "This will be deleted."
./gydnc create subdir/simple-delete-test --title "Delete Me" --body "This will be deleted."
# Use find for a more stable directory listing
find .gydnc/ -print | sort

# Act: delete the entity
# These will output "No matching entities found..." to stdout as per test logs
./gydnc delete simple-delete-test -f
./gydnc delete subdir/simple-delete-test -f

# Assert: list should still show the entities
./gydnc list