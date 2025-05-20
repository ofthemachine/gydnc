#!/bin/bash
set -euo pipefail

./gydnc init . > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Single Entity" --description "A test entity" --tags "test,one" test-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc list --json