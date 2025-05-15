#!/bin/bash
set -e

# Initialize gydnc (assumes localfs backend by default)
./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

CONFIG_FILE="./config.yml"

# Create a sample guidance entity
./gydnc create --config "${CONFIG_FILE}" test-alpha --title "Test Alpha Title" --description "Description for Alpha." --tags "go,test,alpha" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Add some body content manually to the created file for the test
cat << EOF >> ./.gydnc/test-alpha.g6e
This is the body content
for Test Alpha.
It has multiple lines.
EOF

# Execute the command to be tested
./gydnc get --config "${CONFIG_FILE}" test-alpha --output json-frontmatter