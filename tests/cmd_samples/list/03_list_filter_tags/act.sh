#!/bin/bash
set -euo pipefail

# Initialize the repository
./gydnc init . > /dev/null 2>&1

# Create entities with different tags
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Code Entity" --description "A code-related entity" --tags "scope:code,quality:safety" code-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Docs Entity" --description "A documentation entity" --tags "scope:docs,quality:clarity" docs-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Deprecated Entity" --description "A deprecated entity" --tags "scope:code,deprecated" deprecated-entity > /dev/null 2>&1

# List entities with the scope:code tag using --filter-tags
GYDNC_CONFIG=.gydnc/config.yml ./gydnc list --json --filter-tags "scope:code"
