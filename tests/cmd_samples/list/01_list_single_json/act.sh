#!/bin/bash
set -euo pipefail

# Setup: Ensure shared config and data are linked or copied if necessary
# For this test, we assume a config file named 'config.yml' is present
# in the test directory, pointing to 'test_data/' for the 'primary' backend.

./gydnc init . > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Single Entity" --description "A test entity" --tags "test,one" test-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc list