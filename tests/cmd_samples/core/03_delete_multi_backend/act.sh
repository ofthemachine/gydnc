#!/bin/bash
set -euo pipefail

# Arrange: create two backends and the same alias in both
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

./gydnc init --config config.yml
./gydnc create multi-backend-delete --title "Delete Me" --body "In be1" --backend be1 --config config.yml
./gydnc create multi-backend-delete --title "Delete Me" --body "In be2" --backend be2 --config config.yml

# Act: delete the entity (should find in both backends)
./gydnc delete multi-backend-delete -f --config config.yml

# Assert: list should not show the entity in either backend
./gydnc list --config config.yml