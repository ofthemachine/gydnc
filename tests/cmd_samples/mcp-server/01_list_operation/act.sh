#!/bin/bash
set -e

./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

CONFIG_FILE=".gydnc/config.yml"

# Create a test entity
./gydnc create --config "${CONFIG_FILE}" test/mcp-list-test --title "MCP List Test" --tags "test,mcp" >/dev/null 2>&1 || { echo 'create failed'; exit 1; }

# Test MCP server list operation via JSON-RPC
(
  printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n'
  sleep 0.5
  printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gydnc_read","arguments":{"operation":"list"}}}\n'
  sleep 0.5
) | timeout 5 ./gydnc --config "${CONFIG_FILE}" mcp-server 2>&1


