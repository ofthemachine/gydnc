#!/bin/bash
set -e

# Set config for all commands in this script
export GYDNC_CONFIG=./config.yml

# Create a new guidance file using the explicit backend flag
./gydnc create --backend secondary multi_backend/be_flag_test_entity --title "Backend Flag Test"