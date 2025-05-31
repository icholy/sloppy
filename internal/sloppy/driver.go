package sloppy

import (
	"context"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
)

type AgentV2 interface {
	Run(ctx context.Context, input *RunInput) (*RunOutput, error)
}

type RunInput struct {
	Meta           map[string]any
	Prompt         string
	CallToolResult *mcp.CallToolResult
	Tools          []mcp.Tool
}

type RunOutput struct {
	Meta            map[string]any
	CallToolRequest *mcp.CallToolRequest
}

type Driver struct {
	Tools map[string]Tool
	Agent AgentV2
}

func (d *Driver) Loop(ctx context.Context, prompt string) error {
	input := &RunInput{Prompt: prompt}
	for {
		output, err := d.Agent.Run(ctx, input)
		if err != nil {
			return err
		}
		if req := output.CallToolRequest; req != nil {
			res, err := d.call(ctx, *req)
			if err != nil {
				return err
			}
			input = &RunInput{
				CallToolResult: res,
				Meta:           output.Meta,
			}
			continue
		}
		break
	}
	return nil
}

func (d *Driver) call(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tool, ok := d.Tools[req.Method]
	if !ok {
		return mcpx.NewToolResultErrorf("tool not found: %q", req.Method), nil
	}
	return tool.Client.CallTool(ctx, req)
}
