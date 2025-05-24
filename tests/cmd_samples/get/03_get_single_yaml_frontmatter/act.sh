#!/bin/bash
set -e

CONFIG_FILE="./config.yml"
export GYDNC_CONFIG=${CONFIG_FILE} # Use the test-specific config

# Ensure the backend storage path exists (as defined in config.yml)
mkdir -p .gydnc

# Create a sample guidance entity
# Output is suppressed as it's not part of the main assertion for 'get'
./gydnc create test-alpha --title "Test Alpha Title" --description "Description for Alpha." --tags "go,test,alpha" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Add some body content manually to the created file for the test
# The .gydnc path is relative to the temp test dir where config.yml's primary backend points
cat << EOF >> ./.gydnc/test-alpha.g6e
This is the body content
for Test Alpha.
It has multiple lines.
EOF

# Execute the command to be tested
./gydnc get test-alpha # Relies on GYDNC_CONFIG