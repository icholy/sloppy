package sloppy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"github.com/icholy/sloppy/internal/termcolor"
	"github.com/mark3labs/mcp-go/mcp"
)

type Options struct {
	Name   string
	Client *anthropic.Client
	Output io.Writer
	Tools  []Tool
}

type Agent struct {
	name     string
	client   *anthropic.Client
	output   io.Writer
	tools    map[string]Tool
	messages []anthropic.MessageParam
	pending  []anthropic.ContentBlockUnion
}

func New(opt Options) *Agent {
	if opt.Name == "" {
		opt.Name = "Sloppy"
	}
	if opt.Client == nil {
		client := anthropic.NewClient()
		opt.Client = &client
	}
	if opt.Output == nil {
		opt.Output = os.Stdout
	}
	tools := map[string]Tool{}
	for _, tool := range opt.Tools {
		tools[tool.Name] = tool
	}
	return &Agent{
		name:   opt.Name,
		client: opt.Client,
		output: opt.Output,
		tools:  tools,
	}
}

func (a *Agent) Run(ctx context.Context, input *RunInput) (*RunOutput, error) {
	if input.Prompt != "" {
		a.append(anthropic.NewUserMessage(anthropic.NewTextBlock(input.Prompt)))
	}
	if res := input.CallToolResult; res != nil {
		toolUseID, ok := input.Meta["toolUseID"].(string)
		if !ok {
			return nil, fmt.Errorf("missing toolUseId in metadata")
		}
		results := a.toAnthropic(toolUseID, res)
		a.append(anthropic.NewUserMessage(results...))
	}
	if len(a.pending) == 0 {
		response, err := a.llm(ctx, true)
		if err != nil {
			return nil, err
		}
		a.append(response.ToParam())
		a.pending = response.Content
	}
	for len(a.pending) > 0 {
		block := a.pending[0]
		a.pending = a.pending[1:]
		switch block.Type {
		case "text":
			fmt.Fprintf(a.output, "%s: %s\n", termcolor.Text(a.name, termcolor.Yellow), block.Text)
		case "tool_use":
			req, err := a.toMCP(block)
			if err != nil {
				return nil, err
			}
			return &RunOutput{
				CallToolRequest: req,
				Meta:            map[string]any{"toolUseID": block.ID},
			}, nil
		}
	}
	return &RunOutput{}, nil
}

func (a *Agent) LastMessageJSON() string {
	if len(a.messages) == 0 {
		return ""
	}
	last := a.messages[len(a.messages)-1]
	data, _ := json.Marshal(last)
	return string(data)
}

func (a *Agent) toAnthropic(toolUseID string, res *mcp.CallToolResult) []anthropic.ContentBlockParamUnion {
	var results []anthropic.ContentBlockParamUnion
	for _, c := range res.Content {
		if text, ok := c.(mcp.TextContent); ok {
			results = append(results, anthropic.NewToolResultBlock(toolUseID, text.Text, res.IsError))
		} else {
			// TODO: figure out what to do here
			results = append(results, anthropic.NewToolResultBlock(toolUseID, "unsupported response type", true))
		}
	}
	return results
}

func (a *Agent) toMCP(block anthropic.ContentBlockUnion) (*mcp.CallToolRequest, error) {
	var req mcp.CallToolRequest
	req.Params.Name = block.Name
	if err := json.Unmarshal(block.Input, &req.Params.Arguments); err != nil {
		return nil, err
	}
	return &req, nil
}

func (a *Agent) llm(ctx context.Context, tools bool) (*anthropic.Message, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_20250514,
		MaxTokens: 1024,
		Messages:  a.messages,
	}
	if tools {
		for _, tool := range a.tools {
			params.Tools = append(params.Tools, anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        tool.Name,
					Description: anthropic.String(tool.Tool.Description),
					InputSchema: anthropic.ToolInputSchemaParam{
						Type:       constant.Object(tool.Tool.InputSchema.Type),
						Properties: tool.Tool.InputSchema.Properties,
						ExtraFields: map[string]any{
							"required": tool.Tool.InputSchema.Required,
						},
					},
				},
			})
		}
	}
	return a.client.Messages.New(ctx, params)
}

func (a *Agent) append(m anthropic.MessageParam) {
	a.messages = append(a.messages, m)
}
