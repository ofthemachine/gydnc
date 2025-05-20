#!/bin/bash
set -euo pipefail

./gydnc init
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"

cat <<EOF | ./gydnc create my-entity
This is the body.
With a newline.
EOF

./gydnc get my-entity