#!/bin/bash
set -e

# Set config for all commands in this script
export GYDNC_CONFIG=./config.yml

# Initialize gydnc (assumes localfs backend by default)
./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

# Create a sample guidance entity
./gydnc create test-alpha --title "Test Alpha Title" --description "Description for Alpha." --tags "go,test,alpha" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Update the body using the CLI (pipe to update)
cat << EOF | ./gydnc update test-alpha >/dev/null
# Test Alpha Title

Guidance content for 'Test Alpha Title' goes here.
This is the body content
for Test Alpha.
It has multiple lines.
EOF

# Execute the command to be tested
./gydnc get test-alpha --output body