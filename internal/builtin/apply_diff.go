package builtin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/icholy/fuzzypatch"
	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ApplyDiff struct {
	V2        bool
	Threshold float64
}

func (ad *ApplyDiff) ServerToolV2() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("apply_diff",
			mcp.WithDescription(strings.Join([]string{
				"Perform a SEARCH/REPLACE lines operation on a text file.",
				"The search text MUST MATCH FULL LINES.",
				"The search text must match EXACTLY including whitespace and trailing newlines.",
			}, "\n")),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the target file (relative to CWD)."),
			),
			mcp.WithString("search",
				mcp.Required(),
				mcp.Description("Full line contents to search for. Partial line matches will not work."),
			),
			mcp.WithString("replace",
				mcp.Required(),
				mcp.Description("The replacement text."),
			),
			mcp.WithNumber("line",
				mcp.Required(),
				mcp.Description("The line number where the search text starts"),
			),
		),
		Handler: ad.HandleV2,
	}
}

func (ad *ApplyDiff) ServerTool() server.ServerTool {
	if ad.V2 {
		return ad.ServerToolV2()
	}
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

func (ad *ApplyDiff) HandleV2(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path    string  `param:"path,required"`
		Search  string  `param:"search,required"`
		Replace string  `param:"replace,required"`
		Line    float64 `param:"line,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	if input.Search == "" {
		return nil, fmt.Errorf("Search cannot be empty")
	}
	diff := fuzzypatch.Diff{
		Line:    int(input.Line),
		Search:  input.Search,
		Replace: input.Replace,
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, err
	}
	src := string(data)
	edit, ok := fuzzypatch.Search(src, diff, ad.Threshold)
	if !ok {
		return nil, fmt.Errorf("no match for search text: %s", input.Search)
	}
	updated, err := fuzzypatch.Apply(src, []fuzzypatch.Edit{edit})
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(input.Path, []byte(updated), 0644); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("File updated"), nil
}

func (ad *ApplyDiff) Handle(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path string `param:"path,required"`
		Diff string `param:"diff,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	diffs, err := fuzzypatch.Parse(input.Diff)
	if err != nil {
		return nil, err
	}
	if len(diffs) == 0 {
		return nil, fmt.Errorf("no diffs were provided in the request")
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, err
	}
	src := string(data)
	var edits []fuzzypatch.Edit
	for _, d := range diffs {
		e, ok := fuzzypatch.Search(src, d, ad.Threshold)
		if !ok {
			return nil, fmt.Errorf("no match for search text: %s", d.Search)
		}
		edits = append(edits, e)
	}
	updated, err := fuzzypatch.Apply(src, edits)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(input.Path, []byte(updated), 0644); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("File updated"), nil
}
