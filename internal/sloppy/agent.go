package sloppy

import (
	"bufio"
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
	Client *anthropic.Client
	Input  io.Reader
	Output io.Writer
	Tools  []Tool
}

type Agent struct {
	client   *anthropic.Client
	scanner  *bufio.Scanner
	output   io.Writer
	tools    map[string]Tool
	messages []anthropic.MessageParam
}

func New(opt Options) *Agent {
	if opt.Client == nil {
		client := anthropic.NewClient()
		opt.Client = &client
	}
	if opt.Input == nil {
		opt.Input = os.Stdin
	}
	if opt.Output == nil {
		opt.Output = os.Stdout
	}
	tools := map[string]Tool{}
	for _, tool := range opt.Tools {
		tools[tool.Tool.Name] = tool
	}
	return &Agent{
		client:  opt.Client,
		scanner: bufio.NewScanner(opt.Input),
		output:  opt.Output,
		tools:   tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	fmt.Fprintf(a.output, "Tell sloppy what to do\n")
	for {
		fmt.Fprintf(a.output, "%s: ", termcolor.Text("You", termcolor.Blue))
		if !a.scanner.Scan() {
			break
		}
		a.append(anthropic.NewUserMessage(anthropic.NewTextBlock(a.scanner.Text())))
		if err := a.loop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) append(m anthropic.MessageParam) {
	a.messages = append(a.messages, m)
}

func (a *Agent) loop(ctx context.Context) error {
	for {
		response, err := a.llm(ctx)
		if err != nil {
			return err
		}
		a.append(response.ToParam())
		var results []anthropic.ContentBlockParamUnion
		for _, block := range response.Content {
			switch block.Type {
			case "text":
				fmt.Fprintf(a.output, "%s: %s\n", termcolor.Text("Sloppy", termcolor.Yellow), block.Text)
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

func (a *Agent) tool(ctx context.Context, block anthropic.ContentBlockUnion) []anthropic.ContentBlockParamUnion {
	tool, ok := a.tools[block.Name]
	if !ok {
		return []anthropic.ContentBlockParamUnion{
			anthropic.NewToolResultBlock(block.ID, "tool not found", true),
		}
	}
	fmt.Fprintf(a.output, "%s: %s(%s)\n", termcolor.Text("tool", termcolor.Green), block.Name, block.Input)
	var req mcp.CallToolRequest
	req.Params.Name = block.Name
	if err := json.Unmarshal(block.Input, &req.Params.Arguments); err != nil {
		return []anthropic.ContentBlockParamUnion{
			anthropic.NewToolResultBlock(block.ID, err.Error(), true),
		}
	}
	var blocks []anthropic.ContentBlockParamUnion
	res, err := tool.Client.CallTool(ctx, req)
	if err != nil {
		blocks = append(blocks, anthropic.NewToolResultBlock(block.ID, err.Error(), true))
	} else {
		for _, c := range res.Content {
			if text, ok := c.(mcp.TextContent); ok {
				blocks = append(blocks, anthropic.NewToolResultBlock(block.ID, text.Text, res.IsError))
			} else {
				blocks = append(blocks, anthropic.NewToolResultBlock(block.ID, "unsupported response type", true))
			}
		}
	}
	for _, b := range blocks {
		if r := b.OfRequestToolResultBlock; r != nil && r.IsError.Value {
			for _, c := range r.Content {
				var text string
				if t := c.GetText(); t != nil {
					text = *t
				} else {
					data, _ := r.MarshalJSON()
					text = string(data)
				}
				fmt.Fprintf(a.output, "%s: %s", termcolor.Text("error", termcolor.Red), text)
			}
		}
	}
	return blocks
}

func (a *Agent) llm(ctx context.Context) (*anthropic.Message, error) {
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: 1024,
		Messages:  a.messages,
	}
	for _, tool := range a.tools {
		params.Tools = append(params.Tools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Tool.Name,
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
	return a.client.Messages.New(ctx, params)
}
