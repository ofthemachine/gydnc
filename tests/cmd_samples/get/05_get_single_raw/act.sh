#!/bin/bash
set -e

# Initialize gydnc (assumes localfs backend by default)
./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

CONFIG_FILE="./config.yml"

# Create a sample guidance entity (without manually adding body to test raw output of create)
# The .g6e file from create will have:
# ---
# id: test-alpha
# title: Test Alpha Title
# description: Description for Alpha.
# tags:
# - go
# - test
# - alpha
# ---
# # Test Alpha Title
#
# Guidance content for 'Test Alpha Title' goes here.
./gydnc create --config "${CONFIG_FILE}" test-alpha --title "Test Alpha Title" --description "Description for Alpha." --tags "go,test,alpha" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Execute the command to be tested
./gydnc get --config "${CONFIG_FILE}" test-alpha --output raw