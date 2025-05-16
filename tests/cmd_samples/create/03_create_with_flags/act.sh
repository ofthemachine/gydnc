#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory
./gydnc init .

# Create a new guidance file with flags
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Flagged Guidance" --description "Created with flags" --tags "flag,cli" flagged-guidance

# Optional: Display the created file for debugging (not asserted by default)
echo "--- Content of .gydnc/flagged-guidance.g6e: ---"
cat .gydnc/flagged-guidance.g6e