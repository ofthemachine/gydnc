#!/bin/bash
set -euo pipefail

# Arrange: create a guidance entity
./gydnc init
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"
./gydnc create simple-delete-test --title "Delete Me" --body "This will be deleted."
./gydnc create subdir/simple-delete-test --title "Delete Me" --body "This will be deleted."
tree .gydnc/

# Act: delete the entity
./gydnc delete simple-delete-test -f
./gydnc delete subdir/simple-delete-test -f

# Assert: list should not show the entity
./gydnc list