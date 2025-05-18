package builtin

import (
	"context"

	"github.com/icholy/sloppy/internal/sloppy"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolProvider interface {
	ServerTool() server.ServerTool
}

func NewClient(name string, providers ...ToolProvider) (*client.Client, error) {
	server := server.NewMCPServer(
		name,
		"0.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	for _, p := range providers {
		server.AddTools(p.ServerTool())
	}
	return client.NewInProcessClient(server)
}

func Tools(name string, providers ...ToolProvider) []sloppy.Tool {
	client, err := NewClient(name, providers...)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if _, err := client.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		panic(err)
	}
	tools, err := sloppy.ListClientTools(ctx, name, client)
	if err != nil {
		panic(err)
	}
	return tools
}
