#!/bin/bash
set -e

mkdir -p .the_only_store

# Copy the test-specific config.yml into the temp directory
cp ../../multi_backend_test_configs/single_backend_config.yml config.yml

# Create a new guidance file using the only available backend
GYDNC_CONFIG=config.yml ./gydnc create multi_backend/single_be_test_entity --title "Single Backend Test"