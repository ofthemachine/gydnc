#!/bin/bash
# Exit on first error is NOT set here, as we expect the second create to fail

# Initialize gydnc
./gydnc init .

# Create a guidance file for the first time (should succeed)
./gydnc create existing-guidance

# Attempt to create the same guidance file again (should fail)
./gydnc create existing-guidance

# Capture the exit code of the second attempt
SECOND_CREATE_EXIT_CODE=$?
echo "Second create attempt exit code: $SECOND_CREATE_EXIT_CODE"

# Exit with the captured code so assert.yml can check it, but also ensure it's non-zero for the test runner
exit $SECOND_CREATE_EXIT_CODE