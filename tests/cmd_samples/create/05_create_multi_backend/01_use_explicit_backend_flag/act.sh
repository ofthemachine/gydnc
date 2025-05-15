#!/bin/bash
set -e

# Ensure potential store directories are present for the test
# The create command should handle subdirectories within these paths.
mkdir -p .store_primary
mkdir -p .store_secondary

# Run the create command, targeting the 'secondary' backend via flag,
# using the specific config file for this test.
# No init is run to preserve the test-specific config.
./gydnc create --config ./config.yml --backend secondary multi_backend/be_flag_test_entity --title "Backend Flag Test"