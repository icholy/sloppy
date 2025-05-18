package builtin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/icholy/sloppy/internal/patcher"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ApplyDiff struct {
	Threshold float64
}

func (ad *ApplyDiff) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("apply_diff",
			mcp.WithDescription(strings.Join([]string{
				"Apply one or more SEARCH/REPLACE diff blocks to a text file.",
				"",
				"**Block syntax:**",
				"",
				"```",
				"<<<<<<< SEARCH line:<n>",
				"[search text...]",
				"=======",
				"[replace text...]",
				">>>>>>> REPLACE",
				"```",
				"",
				"- You may concatenate multiple blocks in the `diff` parameter.",
				"- The line:n must contain the line number the search text starts at.",
			}, "\n")),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the target file (relative to CWD)."),
			),
			mcp.WithString("diff",
				mcp.Required(),
				mcp.Description("One or more diff blocks in the format above."),
			),
		),
		Handler: ad.Handle,
	}
}

func (ad *ApplyDiff) Handle(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path string `param:"path,required"`
		Diff string `param:"diff,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	diffs, err := patcher.Parse(input.Diff)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, err
	}
	src := string(data)
	var edits []patcher.Edit
	for _, d := range diffs {
		e, ok := patcher.Search(src, d, ad.Threshold)
		if !ok {
			return nil, fmt.Errorf("no match for search text: %s", d.Search)
		}
		edits = append(edits, e)
	}
	updated, err := patcher.Apply(src, edits)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(input.Path, []byte(updated), 0644); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("File updated"), nil
}
