package sloppy

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Tool struct {
	Alias  string
	Tool   mcp.Tool
	Client *client.Client
}

func (t Tool) ToAlias() mcp.Tool {
	tool := t.Tool
	tool.Name = t.Alias
	return tool
}

func ListClientTools(ctx context.Context, name string, c *client.Client) ([]Tool, error) {
	var tools []Tool
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	for _, t := range res.Tools {
		tools = append(tools, Tool{
			Alias:  fmt.Sprintf("%s-%s", name, t.Name),
			Tool:   t,
			Client: c,
		})
	}
	return tools, nil
}
