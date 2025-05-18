#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory
./gydnc init .

# Set config for all commands after init
export GYDNC_CONFIG=.gydnc/config.yml

# Create a new guidance file
./gydnc create existing-guidance

# Try to create the same file again (should fail)
set +e # Allow failure for this command
./gydnc create existing-guidance
SECOND_EXIT_CODE=$?
set -e

echo "Second create attempt exit code: $SECOND_EXIT_CODE"

# Exit with the captured code so assert.yml can check it, but also ensure it's non-zero for the test runner
exit $SECOND_EXIT_CODE