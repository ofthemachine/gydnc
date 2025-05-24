#!/bin/bash
# This script attempts to run 'gydnc list' in an explicitly unconfigured state.
# It is expected to fail or output specific messages indicating no config/backend.

# Ensure no GYDNC environment variables are set, which could point to an existing config
unset GYDNC_CONFIG
unset GYDNC_CONFIG_DIR
unset GYDNC_DEFAULT_BACKEND_NAME
# Add any other relevant GYDNC_... vars if they exist

# The test harness will capture stdout, stderr, and exit code.
./gydnc list