package builtin

import (
	"context"
	"os"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type WriteFile struct{}

func (wf *WriteFile) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("write_file",
			mcp.WithDescription("Write content to a file, replacing its contents or creating it if it doesn't exist."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("The path of the file relative to the current working directory."),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("The content to write to the file."),
			),
		),
		Handler: wf.Handle,
	}
}

func (wf *WriteFile) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path    string `param:"path,required"`
		Content string `param:"content,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to parse arguments", err), nil
	}
	if input.Path == "" {
		return mcp.NewToolResultError("invalid input: path is required"), nil
	}
	if err := os.WriteFile(input.Path, []byte(input.Content), 0644); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to write file", err), nil
	}
	return mcp.NewToolResultText("File written"), nil
}
