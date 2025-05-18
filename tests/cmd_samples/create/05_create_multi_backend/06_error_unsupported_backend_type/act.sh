#!/bin/bash
set -e # Exit on first error

# Use the local config.yml directly

# Try to create a new guidance file with an unsupported backend type (should fail)
set +e # Allow failure for this command
GYDNC_CONFIG=config.yml ./gydnc create multi_backend/unsupported_type_test_entity
EXIT_CODE=$?
set -e

echo "Create attempt exit code: $EXIT_CODE"
exit $EXIT_CODE