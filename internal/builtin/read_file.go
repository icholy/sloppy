package builtin

import (
	"context"
	"os"
	"strings"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ReadFile struct{}

func (rf *ReadFile) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("read_file",
			mcp.WithDescription("Read lines from a file, optionally specifying a start and end line (1-based, inclusive). Returns the file content as a string."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("The path of the file relative to the current working directory."),
			),
			mcp.WithNumber("start_line",
				mcp.Description("The 1-based line number to start reading from (inclusive). If not specified, starts from the first line."),
			),
			mcp.WithNumber("end_line",
				mcp.Description("The 1-based line number to end reading at (inclusive). If not specified, reads to the end of the file."),
			),
		),
		Handler: rf.Handle,
	}
}

func (rf *ReadFile) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path      string  `param:"path,required"`
		StartLine float64 `param:"start_line"`
		EndLine   float64 `param:"end_line"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to parse arguments", err), nil
	}
	if input.Path == "" {
		return mcp.NewToolResultError("invalid input: path is required"), nil
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to read file", err), nil
	}
	lines := strings.SplitAfter(string(data), "\n")
	nlines := len(lines)
	start := 1
	if input.StartLine > 0 {
		start = int(input.StartLine) // use caller‑supplied value
	}
	end := nlines
	if input.EndLine > 0 {
		end = int(input.EndLine) // use caller‑supplied value
	}
	if start < 1 || start > nlines || end < start || end > nlines {
		return mcpx.NewToolResultErrorf("invalid line range %d–%d (file has %d lines)", start, end, nlines), nil
	}
	content := strings.Join(lines[start-1:end], "")
	return mcp.NewToolResultText(content), nil
}
