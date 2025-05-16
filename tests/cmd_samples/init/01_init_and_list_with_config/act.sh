#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory (which will be the temp test directory)
# This will now create .gydnc/config.yml and .gydnc/TAG_ONTOLOGY.md
./gydnc init .

# Run 'gydnc list' using the generated config file
# The config config.yml should now be in .gydnc/ after 'init .'
./gydnc --config .gydnc/config.yml list