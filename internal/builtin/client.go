package builtin

import (
	"context"

	"github.com/icholy/sloppy/internal/sloppy"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewClient returns an in-process MCP server
// which implements the built-in sloppy tools
func NewClient() (*client.Client, error) {
	server := server.NewMCPServer(
		"Sloppy Built-In Tools",
		"0.0.1",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	runCommandTool := &RunCommandTool{}
	editFileTool := &EditFileTool{}
	server.AddTools(
		runCommandTool.ServerTool(),
		editFileTool.ServerTool(),
	)
	return client.NewInProcessClient(server)
}

// Tools returns the built-in tools
func Tools() []sloppy.Tool {
	client, err := NewClient()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if _, err := client.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		panic(err)
	}
	tools, err := sloppy.ListClientTools(ctx, client)
	if err != nil {
		panic(err)
	}
	return tools
}
