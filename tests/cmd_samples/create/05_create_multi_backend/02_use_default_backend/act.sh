#!/bin/bash
set -e

# Ensure potential store directories are present for the test
mkdir -p .store_primary
mkdir -p .store_secondary

# Run the create command, relying on default_backend from config.yml
./gydnc create --config ./config.yml multi_backend/default_be_test_entity --title "Default Backend Test"