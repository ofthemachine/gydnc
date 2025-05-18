#!/bin/bash
set -e # Exit on first error

SUBDIR_NAME="subdir_cfg_create"

# Create the subdirectory if it doesn't exist
mkdir -p "${SUBDIR_NAME}"

# Initialize gydnc in the subdirectory
# This will create SUBDIR_NAME/.gydnc/
# and SUBDIR_NAME/config.yml
./gydnc init "${SUBDIR_NAME}"

# Create a new guidance file using the config file in the subdirectory
# This will now be created at SUBDIR_NAME/.gydnc/my-cfg-guidance.g6e
./gydnc create --config "${SUBDIR_NAME}/.gydnc/config.yml" my-cfg-guidance

# Optional: Display the created file for debugging (not asserted by default)
echo "--- Content of ${SUBDIR_NAME}/.gydnc/my-cfg-guidance.g6e: ---"
cat "${SUBDIR_NAME}/.gydnc/my-cfg-guidance.g6e"

echo "--- Listing ${SUBDIR_NAME}/.gydnc directory for debug: ---"
ls -la "${SUBDIR_NAME}/.gydnc"