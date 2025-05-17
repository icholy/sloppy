package sloppy

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Tool struct {
	Tool   mcp.Tool
	Client *client.Client
}

func ListClientTools(ctx context.Context, c *client.Client) ([]Tool, error) {
	var tools []Tool
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	for _, t := range res.Tools {
		tools = append(tools, Tool{
			Tool:   t,
			Client: c,
		})
	}
	return tools, nil
}
