# Gydnc Architecture Notes

This document tracks architectural decisions and implementation details for the `gydnc` tool.

## Initial Setup

- **Project Structure:** Based on `my_new_ai_plan.md`, section 9.1.
- **Makefile:** Copied from `.agent/service/Makefile`, `BINARY_NAME` changed to `gydnc`.
- **Go Modules:** Not yet initialized.

## Commands

### `llm`
- **Purpose:** Display LLM interaction guidelines.
- **Implementation:** Based on `.agent/service/cmd/llm.go`.
- **Content:** Stored in `gydnc/llm.txt` (embedded). Content adapted from `.agent/service/cmd/llm_cli_help.txt` to reflect `gydnc` commands and concepts.
- **Registration:** Added to `rootCmd` in `gydnc/cmd/root.go`.

## Build System

- **Go Modules:** Initialized with `go mod init gydnc`.
- **Dependencies:** `github.com/spf13/cobra` added via `go mod tidy`.
- **Main Entrypoint:** `gydnc/main.go` created, calls `cmd.Execute()`.