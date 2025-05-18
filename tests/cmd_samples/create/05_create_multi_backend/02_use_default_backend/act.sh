#!/bin/bash
set -e

# Ensure potential store directories are present for the test
mkdir -p .store_primary
mkdir -p .store_secondary

# The test harness copies this test's local config.yml to the temp directory root.
# cp ../../multi_backend_test_configs/default_backend_config.yml config.yml # This line removed

# Create a new guidance file using the default backend
GYDNC_CONFIG=config.yml ./gydnc create multi_backend/default_be_test_entity --title "Default Backend Test"