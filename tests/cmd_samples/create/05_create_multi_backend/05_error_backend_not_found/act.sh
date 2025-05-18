#!/bin/bash
set -e # Exit on first error

# The test harness copies this test's local config.yml to the temp directory root.

# Try to create a new guidance file with a non-existent backend (should fail)
set +e # Allow failure for this command
GYDNC_CONFIG=config.yml ./gydnc create --backend non_existent multi_backend/nonexistent_be_test_entity
EXIT_CODE=$?
set -e

echo "Create attempt exit code: $EXIT_CODE"
exit $EXIT_CODE