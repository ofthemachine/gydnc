#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory (which will be the temp test directory)
./gydnc init .

# Create a new guidance file
# This will now be created at .gydnc/my-new-guidance.g6e
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create my-new-guidance

# Optional: Display the created file for debugging (not asserted by default)
echo "--- Content of my-new-guidance.g6e: ---"
cat .gydnc/my-new-guidance.g6e