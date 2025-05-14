#!/bin/bash
set -e # Exit on first error

# Initialize gydnc (will create .gydnc store, config.yml, TAG_ONTOLOGY.md)
./gydnc init .

# Create a new guidance file with flags
# It should be created in the .gydnc store directly
./gydnc create --title "Flagged Guidance Title" --description "Description from flag" --tags "flag:tag1,category:flag_cat" flagged-guidance

# Optional: Display the created file for debugging
echo "--- Content of .gydnc/flagged-guidance.g6e: ---"
cat .gydnc/flagged-guidance.g6e