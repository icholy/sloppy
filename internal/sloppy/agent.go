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

func (a *Agent) next() (anthropic.ContentBlockUnion, bool) {
	if len(a.pending) == 0 {
		return anthropic.ContentBlockUnion{}, false
	}
	block := a.pending[0]
	a.pending = a.pending[1:]
	return block, true
}

func (a *Agent) Run(ctx context.Context, input *RunInput) (*RunOutput, error) {
	if input.Prompt != "" {
		a.append(anthropic.NewUserMessage(anthropic.NewTextBlock(input.Prompt)))
	}
	if res := input.CallToolResult; res != nil {
		// TODO: how do we get the call id here?
		// need some way to pass around metadata
		results := a.toAnthropicToolResults("", res)
		a.append(anthropic.NewUserMessage(results...))
	}

	a.append(anthropic.NewUserMessage(anthropic.NewTextBlock(input.Prompt)))
	for {
		response, err := a.llm(ctx, true)
		if err != nil {
			return nil, err
		}
		a.append(response.ToParam())
		a.pending = response.Content
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
			return &RunOutput{}, nil
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

func (a *Agent) toAnthropicToolResults(toolUseID string, res *mcp.CallToolResult) []anthropic.ContentBlockParamUnion {
	var results []anthropic.ContentBlockParamUnion
	for _, c := range res.Content {
		if text, ok := c.(mcp.TextContent); ok {
			results = append(results, anthropic.NewToolResultBlock(toolUseID, text.Text, res.IsError))
		} else {
			results = append(results, anthropic.NewToolResultBlock(toolUseID, "unsupported response type", true))
		}
	}
	return results
}

func (a *Agent) toMCPToolRequest(block anthropic.ContentBlockUnion) (*mcp.CallToolRequest, error) {
	var req mcp.CallToolRequest
	req.Params.Name = block.Name
	if err := json.Unmarshal(block.Input, &req.Params.Arguments); err != nil {
		return nil, err
	}
	return &req, nil
}

func (a *Agent) tool(ctx context.Context, block anthropic.ContentBlockUnion) []anthropic.ContentBlockParamUnion {
	tool, ok := a.tools[block.Name]
	if !ok {
		return []anthropic.ContentBlockParamUnion{
			anthropic.NewToolResultBlock(block.ID, "tool not found", true),
		}
	}
	fmt.Fprintf(a.output, "%s: %s(%s)\n", termcolor.Text("tool", termcolor.Green), block.Name, block.Input)
	req, err := a.toMCPToolRequest(block)
	if err != nil {
		return []anthropic.ContentBlockParamUnion{
			anthropic.NewToolResultBlock(block.ID, err.Error(), true),
		}
	}
	var results []anthropic.ContentBlockParamUnion
	res, err := tool.Client.CallTool(ctx, *req)
	if err != nil {
		// TODO: should we just pass this up?
		results = append(results, anthropic.NewToolResultBlock(block.ID, err.Error(), true))
	} else {
		results = a.toAnthropicToolResults(block.ID, res)
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
