#!/bin/bash
set -e # Exit on first error

# Copy the test-specific config.yml into the temp directory
cp ../../multi_backend_test_configs/default_backend_config.yml config.yml

# Try to create a new guidance file with a non-existent backend (should fail)
set +e # Allow failure for this command
GYDNC_CONFIG=config.yml ./gydnc create --backend non_existent multi_backend/nonexistent_be_test_entity
EXIT_CODE=$?
set -e

echo "Create attempt exit code: $EXIT_CODE"