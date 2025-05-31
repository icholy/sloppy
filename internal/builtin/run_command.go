package builtin

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/icholy/sloppy/internal/mcpx"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type RunCommand struct{}

func (rc *RunCommand) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("run_command",
			mcp.WithDescription("Execute a shell command and return its output. Use this for running commands in the terminal."),
			mcp.WithString("command",
				mcp.Required(),
				mcp.Description("The shell command to execute."),
			),
		),
		Handler: rc.Handle,
	}
}

func (rc *RunCommand) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Command string `param:"command,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to parse arguments", err), nil
	}
	if input.Command == "" {
		return mcp.NewToolResultError("invalid arguments: command cannot be empty"), nil
	}

	cmd := exec.Command("bash", "-c", input.Command)
	cmd.Stdin = os.Stdin

	// Both capture and display output
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	cmd.Stderr = io.MultiWriter(os.Stderr, &output)
	if err := cmd.Run(); err != nil {
		return mcpx.NewToolResultErrorf("%v: %s", err, output.String()), nil
	}
	return mcp.NewToolResultText(output.String()), nil
}
