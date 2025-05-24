#!/bin/bash
set -euo pipefail

# Arrange: create a config file and ensure backend dirs exist
cat > config.yml <<EOF
default_backend: be1
storage_backends:
  be1:
    type: localfs
    localfs:
      path: .store_be1
  be2:
    type: localfs
    localfs:
      path: .store_be2
EOF
mkdir -p .store_be1
mkdir -p .store_be2

export GYDNC_CONFIG=./config.yml # Ensure create/delete/list use this config

# Create entities
./gydnc create multi-backend-delete --title "Delete Me" --body "In be1" --backend be1
./gydnc create multi-backend-delete --title "Delete Me" --body "In be2" --backend be2

# Act: delete the entity (should find in both backends)
./gydnc delete multi-backend-delete -f

# Assert: list should not show the entity in either backend
./gydnc list