package sloppy

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
)

type Agent interface {
	Run(ctx context.Context, input *RunInput) (*RunOutput, error)
	LastMessage() string
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
	Root  Agent
	Tools []Tool
	Stack []Agent
}

func (d *Driver) Loop(ctx context.Context, prompt string) error {
	if len(d.Stack) == 0 {
		d.Stack = append(d.Stack, d.Root)
	}
	input := &RunInput{Prompt: prompt}
	for {
		agent := d.Stack[len(d.Stack)-1]
		input.Tools = d.tools()
		output, err := agent.Run(ctx, input)
		if err != nil {
			return err
		}
		if req := output.CallToolRequest; req != nil {
			data, _ := json.MarshalIndent(req.Params, "", "  ")
			fmt.Printf("tool: %s\n", data)

			// we special case the run_agent tool
			if req.Params.Name == "run_agent" {
				var args struct {
					Prompt string `param:"prompt,required"`
					Name   string `param:"name,required"`
				}
				if err := mcpx.MapArguments(req.Params.Arguments, &args); err != nil {
					input = &RunInput{
						Meta:           output.Meta,
						CallToolResult: mcp.NewToolResultErrorFromErr("failed to parse arguments", err),
					}
					continue
				}
				d.Stack = append(d.Stack, NewAnthropicAgent(&AnthropicAgentOptions{Name: args.Name}))
				input = &RunInput{
					Meta: output.Meta,
					Prompt: strings.Join([]string{
						args.Prompt,
						"Note: Only your final response message will be provided back to the user.",
						"This last message should contain all of the relevant information.",
					}, "\n\n"),
				}
			}

			res, err := d.call(ctx, *req)
			if err != nil {
				return err
			}
			data, _ = json.MarshalIndent(res, "", "  ")
			fmt.Printf("output: %s\n", data)
			input = &RunInput{
				CallToolResult: res,
				Meta:           output.Meta,
			}
			continue
		}

		// are we in a nested agent ?
		if len(d.Stack) > 0 {
			input = &RunInput{
				Meta:           map[string]any{}, // uh oh ...
				CallToolResult: mcp.NewToolResultText(agent.LastMessage()),
			}
			d.Stack = d.Stack[:len(d.Stack)-1]
			continue
		}

		break
	}
	return nil
}

func (d *Driver) tools() []mcp.Tool {
	var tools []mcp.Tool
	for _, t := range d.Tools {
		tools = append(tools, t.ToAlias())
	}
	return tools
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
