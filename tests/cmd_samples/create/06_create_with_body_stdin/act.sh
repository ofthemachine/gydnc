#!/bin/bash
set -e
# Initialize gydnc in the current directory, creates ./.gydnc/
./gydnc init . >/dev/null 2>&1 || { echo 'init failed'; exit 1; }
export GYDNC_CONFIG=./.gydnc/config.yml

echo "This is body content from stdin." | ./gydnc create stdin_test --title "Stdin Test Title"
echo "--- Content of .gydnc/stdin_test.g6e: ---"
cat ./.gydnc/stdin_test.g6e