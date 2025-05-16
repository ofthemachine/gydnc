#!/bin/bash
set -e

# Ensure potential store directories are present for the test
mkdir -p .store_primary
mkdir -p .store_secondary

# Copy the test-specific config.yml into the temp directory
cp ../../multi_backend_test_configs/default_backend_config.yml config.yml

# Create a new guidance file using the default backend
GYDNC_CONFIG=config.yml ./gydnc create multi_backend/default_be_test_entity --title "Default Backend Test"