#!/bin/bash
set -e

# Config is provided by harness as config.yml in CWD.
# GYDNC_BIN is ./gydnc in CWD.
# TEST_TEMP_DIR is CWD.

export GYDNC_CONFIG=./config.yml # Use harness-provided config in CWD

ENTITY_ALIAS="update_target_01"
INITIAL_TITLE="Original Title 01"
INITIAL_DESCRIPTION="Original Description"
INITIAL_TAGS="tag1,common"
UPDATED_TITLE="Updated Title for 01"

# Create the initial entity (quietly)
./gydnc create "${ENTITY_ALIAS}" \
  --title "${INITIAL_TITLE}" \
  --description "${INITIAL_DESCRIPTION}" \
  --tags "${INITIAL_TAGS}" >/dev/null 2>&1 # Suppress create's own stdout/stderr

# Update the entity's title (quietly)
./gydnc update "${ENTITY_ALIAS}" --title "${UPDATED_TITLE}" >/dev/null 2>&1 # Suppress update's own stdout/stderr
UPDATE_EXIT_CODE=$?

if [ $UPDATE_EXIT_CODE -ne 0 ]; then
  exit $UPDATE_EXIT_CODE
fi

# Get the updated entity; its JSON output will be asserted and is the ONLY stdout
./gydnc get "${ENTITY_ALIAS}"

exit 0