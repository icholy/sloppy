package builtin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type EditFileTool struct{}

func (t *EditFileTool) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("edit_file",
			mcp.WithDescription("Edit a file in place. To create a file, omit the 'search' parameter."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("The path of the file relative to the current working directory."),
			),
			mcp.WithString("search",
				mcp.Description(strings.Join([]string{
					"Text to search for.",
					"search text must match the associated file section to find EXACTLY.",
					"It must match character-for-character including whitespace, indentation, line endings.",
					"Include all comments, docstrings, etc.",
				}, ". ")),
			),
			mcp.WithString("replace",
				mcp.Description(strings.Join([]string{
					"Text used to replace the the matched 'search' text",
					"will ONLY replace the first match occurrence",
					"Include *just* enough lines in the 'search' parameter to uniquely match each set of lines that need to change",
				}, ". ")),
			),
		),
		Handler: t.Handle,
	}
}

func (t *EditFileTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path   string `param:"path,required"`
		OldStr string `param:"old_str"`
		NewStr string `param:"new_str,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	if input.Path == "" || input.NewStr == input.OldStr {
		return nil, fmt.Errorf("invalid input")
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		if os.IsNotExist(err) && input.OldStr == "" {
			if err := os.WriteFile(input.Path, []byte(input.Path), 0644); err != nil {
				return nil, err
			}
			return mcp.NewToolResultText("File Created"), nil
		}
		return nil, err
	}
	if !bytes.Contains(data, []byte(input.OldStr)) {
		return nil, errors.New("old_str not found in file")
	}
	data = bytes.ReplaceAll(data, []byte(input.OldStr), []byte(input.NewStr))
	if err := os.WriteFile(input.Path, data, 0644); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("File Updated"), nil
}
