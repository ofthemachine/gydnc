#!/bin/bash
set -e

./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

CONFIG_FILE=".gydnc/config.yml"

# Create a test entity with known content
./gydnc create --config "${CONFIG_FILE}" test/mcp-get-test --title "MCP Get Test" --description "Test description" --tags "test,mcp" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Add body content
cat << EOF >> ./.gydnc/test/mcp-get-test.g6e
# MCP Get Test Body

This is the body content for testing the get operation.
EOF

# Test MCP server get operation via JSON-RPC
(
  printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n'
  sleep 0.5
  printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gydnc_read","arguments":{"operation":"get","aliases":["test/mcp-get-test"]}}}\n'
  sleep 0.5
) | timeout 5 ./gydnc --config "${CONFIG_FILE}" mcp-server 2>&1


