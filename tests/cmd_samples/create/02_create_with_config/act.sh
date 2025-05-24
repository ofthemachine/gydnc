#!/bin/bash
set -e # Exit on first error

SUBDIR_NAME="subdir_cfg_create"

# Create the subdirectory if it doesn't exist
mkdir -p "${SUBDIR_NAME}"

# Initialize gydnc in the subdirectory, suppressing output
./gydnc init "${SUBDIR_NAME}" > /dev/null 2>&1

# Create a new guidance file using the config file in the subdirectory
# This will now be created at SUBDIR_NAME/.gydnc/my-cfg-guidance.g6e
./gydnc create --config "${SUBDIR_NAME}/.gydnc/config.yml" my-cfg-guidance

# Verify content using gydnc get
./gydnc get --config "${SUBDIR_NAME}/.gydnc/config.yml" my-cfg-guidance

