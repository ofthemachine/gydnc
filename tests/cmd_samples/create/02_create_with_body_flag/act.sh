#!/bin/bash
set -euo pipefail

./gydnc init > /dev/null
export GYDNC_CONFIG="$(pwd)/.gydnc/config.yml"

# Create entity using --title and --body flags
./gydnc create flag_body_test --title "Flag Body Test" --body "Body from flag" > /dev/null

# Get the entity; this output is checked by assert.yml
./gydnc --output json get flag_body_test