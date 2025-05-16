#!/bin/bash
set -e

# Ensure potential store directories are present for the test
# The create command should handle subdirectories within these paths.
mkdir -p .store_primary
mkdir -p .store_secondary

# Copy the test-specific config.yml into the temp directory
cp ../../multi_backend_test_configs/two_backend_config.yml config.yml

# Create a new guidance file using the explicit backend flag
GYDNC_CONFIG=config.yml ./gydnc create --backend secondary multi_backend/be_flag_test_entity --title "Backend Flag Test"