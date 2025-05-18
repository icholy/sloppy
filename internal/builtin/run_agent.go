package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/icholy/sloppy/internal/mcpx"
	"github.com/icholy/sloppy/internal/sloppy"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type RunAgentTool struct {
	Options *sloppy.Options
	agents  map[string]*sloppy.Agent
}

func (t *RunAgentTool) GetAgent(name string) *sloppy.Agent {
	if t.agents == nil {
		t.agents = map[string]*sloppy.Agent{}
	}
	if a, ok := t.agents[name]; ok {
		return a
	}
	opt := *t.Options
	opt.Name = fmt.Sprintf("Sloppy(%s)", name)
	a := sloppy.New(opt)
	t.agents[name] = a
	return a
}

func (t *RunAgentTool) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("run_agent",
			mcp.WithDescription(strings.Join([]string{
				"Run a child agent to execute tasks.",
				"You may interact with the same agent instance by specifying the same agent name.",
				"Use run_agent to delegate only well-defined, independent sub-tasks.",
				"If you are given a repetitive task, you must create a separate child agent for each repetition.",
				"Do not batch multiple repetitions into a single child.",
				"As the parent agent, coordinate the overall work and only delegate when doing so improves clarity or efficiency.",
			}, " ")),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description(strings.Join([]string{
					"A unique name to identify an agent instance.",
					"This name should be short and descriptive.",
				}, " ")),
			),
			mcp.WithString("prompt",
				mcp.Required(),
				mcp.Description(strings.Join([]string{
					"Instructions for the agent.",
					"Or follow up questions for an agent previously interacted with.",
				}, " ")),
			),
		),
		Handler: t.Handle,
	}
}

var AgentToolSummaryPrompt = strings.Join(
	[]string{
		"Summarize your work since the last user prompt.",
		"This will be the only response the user sees.",
	},
	" ",
)

func (t *RunAgentTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Name   string `param:"name,required"`
		Prompt string `param:"prompt,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return nil, err
	}
	a := t.GetAgent(input.Name)
	if err := a.Run(ctx, input.Prompt, true); err != nil {
		return nil, err
	}
	if err := a.Run(ctx, AgentToolSummaryPrompt, false); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(a.LastMessageJSON()), nil
}
