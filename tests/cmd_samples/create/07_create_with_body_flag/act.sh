#!/bin/bash
set -e
./gydnc init . >/dev/null 2>&1 || { echo 'init failed'; exit 1; }
export GYDNC_CONFIG=./.gydnc/config.yml

./gydnc create flag_test --title "Flag Test Title" --body "This is body content from the --body flag."
echo "--- Content of .gydnc/flag_test.g6e: ---"
cat ./.gydnc/flag_test.g6e