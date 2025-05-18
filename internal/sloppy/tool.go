package sloppy

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Tool struct {
	Name   string
	Tool   mcp.Tool
	Client *client.Client
}

func ListClientTools(ctx context.Context, name string, c *client.Client) ([]Tool, error) {
	var tools []Tool
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	for _, t := range res.Tools {
		tools = append(tools, Tool{
			Name:   fmt.Sprintf("%s-%s", name, t.Name),
			Tool:   t,
			Client: c,
		})
	}
	return tools, nil
}
