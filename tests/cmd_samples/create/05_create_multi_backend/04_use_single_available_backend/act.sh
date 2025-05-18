#!/bin/bash
set -e

mkdir -p .the_only_store

# Create a new guidance file using the only available backend
GYDNC_CONFIG=config.yml ./gydnc create multi_backend/single_be_test_entity --title "Single Backend Test"