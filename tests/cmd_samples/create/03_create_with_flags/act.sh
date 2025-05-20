#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory
./gydnc init . > /dev/null 2>&1 || { echo 'init failed'; exit 1; }

# Create a new guidance file with flags
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Flagged Guidance" --description "Created with flags" --tags "flag,cli" flagged-guidance > /dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Assert on the output of 'gydnc get' instead of file content
GYDNC_CONFIG=.gydnc/config.yml ./gydnc get flagged-guidance