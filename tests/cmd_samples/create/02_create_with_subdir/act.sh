#!/bin/bash
set -e # Exit on first error

# Initialize gydnc in the current directory
./gydnc init .

# Create a new guidance file with a path-like alias
# This should create .gydnc/category/my-sub-guidance.g6e
./gydnc create category/my-sub-guidance

# Optional: Display the created file for debugging
echo "--- Content of .gydnc/category/my-sub-guidance.g6e: ---"
cat .gydnc/category/my-sub-guidance.g6e