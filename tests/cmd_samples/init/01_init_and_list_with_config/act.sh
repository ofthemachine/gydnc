#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory (which will be the temp test directory)
# This will now create config.yml, TAG_ONTOLOGY.md at root, and .gydnc/ store directory.
./gydnc init .

# Run 'gydnc list' using the generated config file
# The config config.yml should be in the current directory (tempDir) after 'init .'
./gydnc --config config.yml list