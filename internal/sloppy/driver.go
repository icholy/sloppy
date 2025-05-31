package sloppy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
)

type Agent interface {
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
	Agent Agent
	Tools []Tool
}

func (d *Driver) Loop(ctx context.Context, prompt string) error {
	input := &RunInput{Prompt: prompt}
	for _, t := range d.Tools {
		input.Tools = append(input.Tools, t.ToAlias())
	}
	for {
		output, err := d.Agent.Run(ctx, input)
		if err != nil {
			return err
		}
		if req := output.CallToolRequest; req != nil {
			data, _ := json.MarshalIndent(req.Params, "", "  ")
			fmt.Printf("tool: %s\n", data)
			res, err := d.call(ctx, *req)
			if err != nil {
				return err
			}
			data, _ = json.MarshalIndent(res, "", "  ")
			fmt.Printf("output: %s\n", data)
			input = &RunInput{CallToolResult: res, Meta: output.Meta}
			for _, t := range d.Tools {
				input.Tools = append(input.Tools, t.ToAlias())
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
		if t.Alias == req.Params.Name {
			found = true
			tool = t
			break
		}
	}
	if !found {
		return mcpx.NewToolResultErrorf("tool not found: %q", req.Params.Name), nil
	}
	// replace the alias name with the actual name before making request
	req.Params.Name = tool.Tool.Name
	return tool.Client.CallTool(ctx, req)
}
