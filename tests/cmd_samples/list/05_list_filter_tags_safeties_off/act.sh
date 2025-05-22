#!/bin/bash
set -euo pipefail

# Initialize the repository
./gydnc init . > /dev/null 2>&1

# Create entities with different quality tags
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Safety Entity" --description "A safety-related entity" --tags "quality:safety,scope:code" safety-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Clarity Entity" --description "A clarity-related entity" --tags "quality:clarity,scope:docs" clarity-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Performance Entity" --description "A performance-related entity" --tags "quality:performance,scope:code" performance-entity > /dev/null 2>&1
GYDNC_CONFIG=.gydnc/config.yml ./gydnc create --title "Non-Quality Entity" --description "A non-quality entity" --tags "feature:awesome" feature-entity > /dev/null 2>&1

# The "safeties off" query - all quality tags except safety
GYDNC_CONFIG=.gydnc/config.yml ./gydnc list --json --filter-tags "quality:* -quality:safety"