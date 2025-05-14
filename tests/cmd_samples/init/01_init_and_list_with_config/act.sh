#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory (which will be the temp test directory)
./gydnc init .

# Run 'gydnc list' using the generated config file
# The test harness copies the gydnc binary into the temp dir.
# The config gydnc.conf should be in the current directory (tempDir) after 'init .'
./gydnc --config gydnc.conf list