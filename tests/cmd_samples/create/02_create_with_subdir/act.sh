#!/bin/bash
set -e # Exit on first error

SUBDIR_NAME="category"

# Initialize gydnc in the subdirectory
./gydnc init "${SUBDIR_NAME}"

# Create a new guidance file using the --config flag
# This will now be created at ${SUBDIR_NAME}/.gydnc/my-sub-guidance.g6e
./gydnc create --config "${SUBDIR_NAME}/.gydnc/config.yml" "${SUBDIR_NAME}/my-sub-guidance"

# Optional: Display the created file for debugging (not asserted by default)
echo "--- Content of ${SUBDIR_NAME}/.gydnc/my-sub-guidance.g6e: ---"
cat "${SUBDIR_NAME}/.gydnc/my-sub-guidance.g6e"