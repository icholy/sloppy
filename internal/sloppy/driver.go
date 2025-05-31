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
	Agent AgentV2
	Tools []Tool
}

func (d *Driver) Loop(ctx context.Context, prompt string) error {
	input := &RunInput{Prompt: prompt}
	for _, t := range d.Tools {
		input.Tools = append(input.Tools, t.Tool)
	}
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
			for _, t := range d.Tools {
				input.Tools = append(input.Tools, t.Tool)
			}
			continue
		}
		break
	}
	return nil
}

func (d *Driver) call(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var found bool
	var tool Tool
	for _, t := range d.Tools {
		if t.Name == req.Method {
			found = true
			tool = t
		}
	}
	if !found {
		return mcpx.NewToolResultErrorf("tool not found: %q", req.Method), nil
	}
	return tool.Client.CallTool(ctx, req)
}
