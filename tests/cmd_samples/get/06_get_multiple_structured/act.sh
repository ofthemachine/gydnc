#!/bin/bash
set -e

# Initialize gydnc
./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

CONFIG_FILE="./config.yml"

# Create first entity: test-alpha
./gydnc create --config "${CONFIG_FILE}" test-alpha --title "Test Alpha Title" --description "Alpha Desc" --tags "alpha,common" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }
cat << EOF >> ./.gydnc/test-alpha.g6e
Body for Alpha.
EOF

# Create second entity: test-beta
./gydnc create --config "${CONFIG_FILE}" test-beta --title "Test Beta Title" --description "Beta Desc" --tags "beta,common" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }
cat << EOF >> ./.gydnc/test-beta.g6e
Body for Beta.
EOF

# Execute the command to be tested
./gydnc get --config "${CONFIG_FILE}" test-alpha test-beta # Default output is 'structured'