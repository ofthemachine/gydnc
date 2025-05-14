#!/bin/bash
# This script attempts to run 'gydnc list' without any prior setup.
# It is expected to fail or output specific messages indicating no config/backend.

# The test harness will capture stdout, stderr, and exit code.
./gydnc list