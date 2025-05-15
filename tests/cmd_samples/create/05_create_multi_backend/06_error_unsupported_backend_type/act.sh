#!/bin/bash
set -e

# Attempt to create with an unsupported backend type
./gydnc create --config ./config.yml multi_backend/unsupported_type_test --title "Unsupported Type Test"