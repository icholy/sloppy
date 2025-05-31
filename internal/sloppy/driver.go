package sloppy

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type AgentV2 interface {
	Run(ctx context.Context, input *RunInput) (*RunOutput, error)
}

type Driver struct {
	a AgentV2
}

func (d *Driver) Run(ctx context.Context, prompt string) error {

	input := &RunInput{
		Prompt: prompt,
	}

	for {
		output, err := d.a.Run(ctx, input)
		if err != nil {
			return err
		}
		if req := output.CallToolRequest; req != nil {
			input = &RunInput{
				CallToolResult: d.call(req),
			}
			continue
		}
		break
	}

	return nil
}

func (d *Driver) call(req *mcp.CallToolRequest) *mcp.CallToolResult {
	return nil
}
