#!/bin/bash
set -e

mkdir -p .store_primary

# Attempt to create with a non-existent backend name
./gydnc create --config ./config.yml --backend non_existent multi_backend/non_existent_be_test --title "NonExistent Test"