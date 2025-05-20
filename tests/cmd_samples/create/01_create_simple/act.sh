#!/bin/bash
set -euo pipefail

# Arrange: initialize and create a simple guidance entity
./gydnc init > /dev/null
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"
cat <<EOF | ./gydnc create my-new-guidance --title "my-new-guidance" > /dev/null
# my-new-guidance

Guidance content for 'my-new-guidance' goes here.
EOF

./gydnc get my-new-guidance