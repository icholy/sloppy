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

func (a *Agent) Run(ctx context.Context, input string, tools bool) error {
	a.append(anthropic.NewUserMessage(anthropic.NewTextBlock(input)))
	for {
		response, err := a.llm(ctx, tools)
		if err != nil {
			return err
		}
		a.append(response.ToParam())
		var results []anthropic.ContentBlockParamUnion
		for _, block := range response.Content {
			switch block.Type {
			case "text":
				fmt.Fprintf(a.output, "%s: %s\n", termcolor.Text(a.name, termcolor.Yellow), block.Text)
			case "tool_use":
				results = append(results, a.tool(ctx, block)...)
			}
		}
		if len(results) == 0 {
			return nil
		}
		a.append(anthropic.NewUserMessage(results...))
	}
}

func (a *Agent) LastMessageJSON() string {
	if len(a.messages) == 0 {
		return ""
	}
	last := a.messages[len(a.messages)-1]
	data, _ := json.Marshal(last)
	return string(data)
}

func (a *Agent) tool(ctx context.Context, block anthropic.ContentBlockUnion) []anthropic.ContentBlockParamUnion {
	tool, ok := a.tools[block.Name]
	if !ok {
		return []anthropic.ContentBlockParamUnion{
			anthropic.NewToolResultBlock(block.ID, "tool not found", true),
		}
	}
	fmt.Fprintf(a.output, "%s: %s(%s)\n", termcolor.Text("tool", termcolor.Green), block.Name, block.Input)
	var req mcp.CallToolRequest
	req.Params.Name = tool.Tool.Name
	if err := json.Unmarshal(block.Input, &req.Params.Arguments); err != nil {
		return []anthropic.ContentBlockParamUnion{
			anthropic.NewToolResultBlock(block.ID, err.Error(), true),
		}
	}
	var results []anthropic.ContentBlockParamUnion
	res, err := tool.Client.CallTool(ctx, req)
	if err != nil {
		// TODO: should we just pass this up?
		results = append(results, anthropic.NewToolResultBlock(block.ID, err.Error(), true))
	} else {
		for _, c := range res.Content {
			if text, ok := c.(mcp.TextContent); ok {
				results = append(results, anthropic.NewToolResultBlock(block.ID, text.Text, res.IsError))
			} else {
				results = append(results, anthropic.NewToolResultBlock(block.ID, "unsupported response type", true))
			}
		}
	}
	for _, b := range results {
		if r := b.OfToolResult; r != nil && r.IsError.Value {
			for _, c := range r.Content {
				var text string
				if t := c.GetText(); t != nil {
					text = *t
				} else {
					data, _ := r.MarshalJSON()
					text = string(data)
				}
				fmt.Fprintf(a.output, "%s: %s\n", termcolor.Text("error", termcolor.Red), text)
			}
		}
	}
	return results
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
