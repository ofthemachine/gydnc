#!/bin/bash
set -e # Exit on first error

# The test harness copies this test's local config.yml to the temp directory root.

# Try to create a new guidance file with ambiguous backend (should fail)
set +e # Allow failure for this command
GYDNC_CONFIG=config.yml ./gydnc create multi_backend/ambiguous_test_entity
EXIT_CODE=$?
set -e

echo "Create attempt exit code: $EXIT_CODE"
exit $EXIT_CODE