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

type EditFile struct{}

func (ef *EditFile) ServerTool() server.ServerTool {
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
		Handler: ef.Handle,
	}
}

func (ef *EditFile) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Path    string `param:"path,required"`
		Search  string `param:"search"`
		Replace string `param:"replace,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	if input.Path == "" || input.Replace == input.Search {
		return nil, fmt.Errorf("invalid input")
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		if os.IsNotExist(err) && input.Search == "" {
			if err := os.WriteFile(input.Path, []byte(input.Path), 0644); err != nil {
				return nil, err
			}
			return mcp.NewToolResultText("File Created"), nil
		}
		return nil, err
	}
	if !bytes.Contains(data, []byte(input.Search)) {
		return nil, errors.New("search text not found in file")
	}
	data = bytes.ReplaceAll(data, []byte(input.Search), []byte(input.Replace))
	if err := os.WriteFile(input.Path, data, 0644); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("File Updated"), nil
}
