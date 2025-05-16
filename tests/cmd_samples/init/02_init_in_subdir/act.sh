#!/bin/bash
set -e # Exit on first error

# Define the subdirectory name
SUBDIR_NAME="subdir"

# Create the subdirectory if it doesn't exist (gydnc init will also do this)
mkdir -p "$SUBDIR_NAME"

# Initialize gydnc in the specified subdirectory
# This will create $SUBDIR_NAME/.gydnc/config.yml and $SUBDIR_NAME/.gydnc/TAG_ONTOLOGY.md
./gydnc init "$SUBDIR_NAME"

echo "--- Verifying with list command --- "
# Run 'gydnc list' using the generated config file in the subdirectory
./gydnc --config "$SUBDIR_NAME/.gydnc/config.yml" list

# Debug output (optional, but can be helpful)
echo "--- Directory structure created in $SUBDIR_NAME: ---"
ls -R "$SUBDIR_NAME"
echo "--- Content of $SUBDIR_NAME/.gydnc/config.yml: ---"
cat "$SUBDIR_NAME/.gydnc/config.yml"