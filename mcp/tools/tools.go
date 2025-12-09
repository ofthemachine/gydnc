package tools

import (
	"gydnc/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var Server *mcp.Server

// AppContext is set by the mcp-server command before starting the server
var AppContext *service.AppContext

func init() {
	Server = mcp.NewServer(
		&mcp.Implementation{
			Name:    "gydnc",
			Title:   "gydnc - Guidance Knowledge Base",
			Version: "v0.0.1",
		}, nil)

	mcp.AddTool(Server, GuidanceReadTool, GuidanceRead)
	mcp.AddTool(Server, GuidanceWriteTool, GuidanceWrite)
}
