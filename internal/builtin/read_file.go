package builtin

import (
	"context"
	"fmt"
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
		Path      string `param:"path,required"`
		StartLine int    `param:"start_line"`
		EndLine   int    `param:"end_line"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	if input.Path == "" {
		return nil, fmt.Errorf("invalid input: path is required")
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, err
	}
	lines := strings.SplitAfter(string(data), "\n")
	nlines := len(lines)
	start := max(1, input.StartLine)
	end := min(len(lines), input.EndLine)
	if start < 1 || start > nlines {
		return nil, fmt.Errorf("start_line %d out of range (file has %d lines)", start, nlines)
	}
	if end < start || end > nlines {
		return nil, fmt.Errorf("end_line %d out of range (file has %d lines)", end, nlines)
	}
	content := strings.Join(lines[start-1:end], "")
	return mcp.NewToolResultText(content), nil
}
