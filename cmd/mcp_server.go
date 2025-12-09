package cmd

import (
	"context"
	"fmt"

	"gydnc/mcp/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// mcpCommandDescription is the help text for the mcp-server command.
// Update this when tools are added/removed to keep documentation in sync.
const mcpCommandDescription = `Run gydnc as a Model Context Protocol (MCP) server over stdio.
This allows AI agents to interact with the gydnc knowledge base through
standardized MCP tool calls. The server exposes the following tools:

- gydnc_read: Read guidance entities (operations: 'list' to discover entities, 'get' to retrieve full content)
- gydnc_write: Write guidance entities (operations: 'create' to add new entities, 'update' to modify existing ones)

The server communicates via JSON-RPC over stdio.`

var mcpServerCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Run gydnc as an MCP server",
	Long:  mcpCommandDescription,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config is initialized (initConfig should have run, but verify)
		if appContext == nil || appContext.Config == nil {
			return fmt.Errorf("application context not initialized; config required for MCP server")
		}

		// Set the AppContext in the tools package so handlers can access it
		tools.AppContext = appContext

		// Run the MCP server with stdio transport
		ctx := context.Background()
		return tools.Server.Run(ctx, &mcp.StdioTransport{})
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(mcpServerCmd)
}
