package sloppy

import (
	"context"
	"encoding/json"
	"fmt"
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

type Frame struct {
	Name  string
	Meta  map[string]any
	Agent Agent
}

type Driver struct {
	Tools    []Tool
	Stack    []Frame
	NewAgent func(name string) Agent
}

func (d *Driver) Loop(ctx context.Context, prompt string) error {
	if len(d.Stack) == 0 {
		d.Stack = append(d.Stack, Frame{
			Name:  "sloppy",
			Agent: d.NewAgent(""),
		})
	}
	input := &RunInput{Prompt: prompt}
	for {
		frame := d.Stack[len(d.Stack)-1]
		agent := frame.Agent
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
				d.Stack = append(d.Stack, Frame{
					Name:  args.Name,
					Meta:  output.Meta,
					Agent: d.NewAgent(args.Name),
				})
				input = &RunInput{
					Meta: output.Meta,
					Prompt: strings.Join([]string{
						args.Prompt,
						"Note: Only your final response message will be provided back to the user.",
						"This last message should contain all of the relevant information.",
					}, "\n\n"),
				}
				continue
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
		if len(d.Stack) > 1 {
			input = &RunInput{
				Meta:           frame.Meta,
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
	return append(tools, mcp.NewTool("run_agent",
		mcp.WithDescription(strings.Join([]string{
			"Run a child agent to execute a sub-task.",
			"You are responsible for planning, breaking down the work, and sequencing.",
			"Do not delegate the entire task to another agent.",
			"You are still the executor, complete the core task yourself.",
			"You MUST HAVE AT LEAST 3 instances of the same template task to use an agent",
			"Create a SEPARATE agent for EACH INSTNACE of a repetaive tasks.",
			"Each SUB-TASK should have a SEPARATELY NAMED agent.",
			"Create NEW AGENTS FOR EACH NEW SUB-TASK.",
			"Do NOT re-use the same agent for multiple tasks.",
			"ONLY re-use agents to ask for corrections or follow up questions to THEIR SPECIFIC TASK.",
		}, " ")),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description(strings.Join([]string{
				"A unique name to identify an agent instance.",
				"This name should be short and descriptive.",
			}, " ")),
		),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description(strings.Join([]string{
				"Instructions for the agent.",
				"Or follow up questions for an agent previously interacted with.",
			}, " ")),
		),
	))
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
