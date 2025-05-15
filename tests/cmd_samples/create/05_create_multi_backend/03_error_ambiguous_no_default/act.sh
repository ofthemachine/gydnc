#!/bin/bash
set -e

mkdir -p .store_primary
mkdir -p .store_secondary

# Attempt to create without specifying backend, expecting an error
# The actual script will exit with non-zero due to `set -e` if gydnc errors, which is caught by harness.
./gydnc create --config ./config.yml multi_backend/ambiguous_test_entity --title "Ambiguous Test"