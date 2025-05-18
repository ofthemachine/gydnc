#!/bin/bash
# set -e # We expect a failure

./gydnc init . >/dev/null 2>&1
export GYDNC_CONFIG=./.gydnc/config.yml

# Try the command. If it succeeds (exit 0), this script will exit 0 (bad for test).
# If it fails (exit non-0), this script will exit with that non-0 code (good for test).
echo "Body from stdin" | ./gydnc create multi_source_test --title "Multi Source Error Test" --body "Body from --body flag"
cmd_exit_code=$?

if [ $cmd_exit_code -ne 0 ]; then
  echo "Command failed as expected (exit code $cmd_exit_code)." # For stdout assertion
  # stderr will be captured by harness directly from gydnc
  exit 1 # Force script exit code to 1 for assert.yml
else
  echo "Command Succeeded, but was expected to fail!" >&2
  exit 0 # Script succeeded, but this means the test assertion for exit_code 1 will fail.
fi