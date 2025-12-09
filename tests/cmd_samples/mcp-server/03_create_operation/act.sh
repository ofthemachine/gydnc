#!/bin/bash
set -e

./gydnc init >/dev/null 2>&1 || { echo 'init failed'; exit 1; }

CONFIG_FILE=".gydnc/config.yml"

# Test MCP server create operation via JSON-RPC
# Use printf to properly handle newlines in JSON strings
(
  printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n'
  sleep 0.5
  printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"gydnc_write","arguments":{"operation":"create","alias":"test/mcp-create-test","title":"MCP Create Test","description":"Created via MCP","tags":["test","mcp"],"body":"# MCP Create Test\\n\\nThis entity was created via MCP server."}}}\n'
  sleep 0.5
) | timeout 5 ./gydnc --config "${CONFIG_FILE}" mcp-server 2>&1

