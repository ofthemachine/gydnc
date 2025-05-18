#!/bin/bash
set -e

# Config is provided by harness as config.yml in CWD.
# GYDNC_BIN is ./gydnc in CWD.
# TEST_TEMP_DIR is CWD.

# Initialize gydnc (quietly) - this creates ./.gydnc/
# and ./.gydnc/config.yml.
# We will use the harness-provided config.yml directly via GYDNC_CONFIG.
./gydnc init > /dev/null 2>&1 # Suppress init's own stdout/stderr

export GYDNC_CONFIG="${TEST_TEMP_DIR}/config.yml" # Use harness-provided config

ENTITY_ALIAS="update_target_01"
INITIAL_TITLE="Original Title 01"
INITIAL_DESCRIPTION="Original Description"
INITIAL_TAGS="tag1,common"
UPDATED_TITLE="Updated Title for 01"

# Create the initial entity (quietly)
./gydnc create "${ENTITY_ALIAS}" \
  --title "${INITIAL_TITLE}" \
  --description "${INITIAL_DESCRIPTION}" \
  --tags "${INITIAL_TAGS}" > /dev/null # Suppress create's own stdout

# Update the entity's title - THIS is the command whose output we want to check
# Capture its stdout and stderr separately
UPDATE_STDOUT_STDERR=$(./gydnc update "${ENTITY_ALIAS}" --title "${UPDATED_TITLE}" 2>&1)
UPDATE_EXIT_CODE=$?

# Output only the line containing "Updated guidance:" from the update command's output
# This will be the *only* stdout from this script for assertion purposes.
echo "${UPDATE_STDOUT_STDERR}" | grep "Updated guidance:" || true
# The '|| true' ensures grep doesn't cause script exit if the line isn't found (though test would fail assertion)


# If update command failed, script should reflect that.
if [ $UPDATE_EXIT_CODE -ne 0 ]; then
  # echo "DEBUG: Update command failed. Output was:" >&2 # Optional debug
  # echo "${UPDATE_STDOUT_STDERR}" >&2 # Optional debug
  exit $UPDATE_EXIT_CODE
fi

exit 0 # Explicitly exit 0 if update was successful