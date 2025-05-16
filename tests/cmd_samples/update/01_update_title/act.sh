#!/bin/bash
set -e

# TEST_TEMP_DIR is CWD, set by harness.
# GYDNC_BIN is ./gydnc, copied into CWD by harness.
# GYDNC_CONFIG is set to "" by the test harness initially for this test.

# Initialize gydnc. This will create ./config.yml in CWD.
echo "Initializing gydnc (stderr will be visible)..."
./gydnc init

# Explicitly set GYDNC_CONFIG to the newly created config file for subsequent commands.
# Use an absolute path. TEST_TEMP_DIR (CWD) is already an absolute path from the harness.
CONFIG_FILE_NAME="config.yml"
ABSOLUTE_CONFIG_PATH="${TEST_TEMP_DIR}/${CONFIG_FILE_NAME}"

if [ ! -f "${ABSOLUTE_CONFIG_PATH}" ]; then
  echo "CRITICAL ERROR: ${ABSOLUTE_CONFIG_PATH} not found after 'gydnc init'!"
  ls -la "${TEST_TEMP_DIR}"
  exit 1
fi

export GYDNC_CONFIG="${ABSOLUTE_CONFIG_PATH}"
echo "Shell GYDNC_CONFIG set to absolute path: ${GYDNC_CONFIG}"

ENTITY_ALIAS="update_target_01"
# Initial title should match what's expected in the body template of the assertion
INITIAL_TITLE="Original Title 01"
# Description and tags for the initial creation, matching assertion's expectation for persisted fields
INITIAL_DESCRIPTION="Original Description"
INITIAL_TAGS="tag1,common" # Create command should sort these to common,tag1
# Updated title, matching assertion's expectation for the updated field
UPDATED_TITLE="Updated Title for 01"

# Create the initial entity with all fields expected by the assertion's final state (except updated title)
echo "Creating entity '${ENTITY_ALIAS}' with initial properties..."
./gydnc create "${ENTITY_ALIAS}" \
  --title "${INITIAL_TITLE}" \
  --description "${INITIAL_DESCRIPTION}" \
  --tags "${INITIAL_TAGS}"

# Add a small delay and an ls to ensure file system changes are flushed/visible
sleep 0.1
echo "Listing store before update:"
ls -l "${TEST_TEMP_DIR}/.gydnc/"

# Update the entity's title
echo "Updating entity '${ENTITY_ALIAS}' title to '${UPDATED_TITLE}'..."
./gydnc update "${ENTITY_ALIAS}" --title "${UPDATED_TITLE}"

# The 'gydnc update' command should output "Updated guidance: .gydnc/update_target_01.g6e"
# which is checked by stdout assertion.

# For final state verification (useful if assertions fail on content):
echo "Final content of .gydnc/${ENTITY_ALIAS}.g6e:"
if [ -f "${TEST_TEMP_DIR}/.gydnc/${ENTITY_ALIAS}.g6e" ]; then
 cat "${TEST_TEMP_DIR}/.gydnc/${ENTITY_ALIAS}.g6e"
else
  echo "ERROR: Entity file ${TEST_TEMP_DIR}/.gydnc/${ENTITY_ALIAS}.g6e not found after update!"
fi