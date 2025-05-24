#!/bin/bash
set -euo pipefail

# Create a config file for the test
cat > .gydnc/config.yml <<EOF
default_backend: backend1
storage_backends:
  backend1:
    type: localfs
    localfs:
      path: .gydnc/backend1
  backend2:
    type: localfs
    localfs:
      path: .gydnc/backend2
EOF

export GYDNC_CONFIG=.gydnc/config.yml

# Create guidance entity in backend1
cat << 'EOF' | ./gydnc create test-entity --title "Test Entity" --backend "backend1"
This is test content that should appear in backend1.
EOF

# Create guidance entity in backend2
cat << 'EOF' | ./gydnc create test-entity --title "Test Entity" --backend "backend2"
This is test content that should appear in backend2.
EOF

# List contents of both backends to verify
echo "=== Listing backend1 contents ==="
./gydnc list --backend backend1

echo "=== Listing backend2 contents ==="
./gydnc list --backend backend2