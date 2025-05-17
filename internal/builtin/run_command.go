package builtin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/icholy/sloppy/internal/mcpx"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type RunCommandTool struct{}

func (t *RunCommandTool) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("run_command",
			mcp.WithDescription("Execute a shell command and return its output. Use this for running commands in the terminal."),
			mcp.WithString("command",
				mcp.Required(),
				mcp.Description("The shell command to execute."),
			),
		),
		Handler: t.Handle,
	}
}

func (t *RunCommandTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Command string `param:"command,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	if input.Command == "" {
		return nil, errors.New("command cannot be empty")
	}

	cmd := exec.Command("bash", "-c", input.Command)
	cmd.Stdin = os.Stdin

	// Both capture and display output
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	cmd.Stderr = io.MultiWriter(os.Stderr, &output)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, output.String())
	}
	return mcp.NewToolResultText(output.String()), nil
}
