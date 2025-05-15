#!/bin/bash
set -e

mkdir -p .the_only_store

# Create when only one backend is defined, no default, no flag
./gydnc create --config ./config.yml multi_backend/single_be_test_entity --title "Single Backend Test"