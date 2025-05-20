#!/bin/bash
set -euo pipefail

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

./gydnc init --config config.yml > /dev/null 2>&1
GYDNC_CONFIG=config.yml ./gydnc create --title "Entity in BE1" --description "From backend 1" --tags "be1,shared" shared-entity --backend be1 --config config.yml > /dev/null 2>&1
GYDNC_CONFIG=config.yml ./gydnc create --title "Entity in BE2" --description "From backend 2" --tags "be2,shared" shared-entity --backend be2 --config config.yml > /dev/null 2>&1
GYDNC_CONFIG=config.yml ./gydnc create --title "Unique in BE2" --description "Unique entity" --tags "be2,unique" unique-entity --backend be2 --config config.yml > /dev/null 2>&1
GYDNC_CONFIG=config.yml ./gydnc list --json --config config.yml