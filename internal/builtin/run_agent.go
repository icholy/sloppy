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

type RunAgent struct {
	Options *sloppy.Options
	agents  map[string]*sloppy.Agent
}

func (ra *RunAgent) GetAgent(name string) *sloppy.Agent {
	if ra.agents == nil {
		ra.agents = map[string]*sloppy.Agent{}
	}
	if a, ok := ra.agents[name]; ok {
		return a
	}
	opt := *ra.Options
	opt.Name = fmt.Sprintf("Sloppy(%s)", name)
	a := sloppy.New(opt)
	ra.agents[name] = a
	return a
}

func (ra *RunAgent) ServerTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("run_agent",
			mcp.WithDescription(strings.Join([]string{
				"Run a child agent to execute a sub-task.",
				"You are responsible for planning, breaking down the work, and sequencing.",
				"Do not delegate the entire task to another agent.",
				"You are still the executor, complete the core task yourself.",
				"You MUST HAVE AT LEAST 3 instances of the same template task to use an agent",
				"Create a SEPARATE agent for EACH INSTNACE of a repetaive tasks.",
				"Each SUB-TASK should have a SEPARATELY NAMED agent.",
				"Create NEW AGENTS FOR EACH NEW SUB-TASK.",
				"Do NOT re-use the same agent for multiple tasks.",
				"ONLY re-use agents to ask for corrections or follow up questions to THEIR SPECIFIC TASK.",
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
		Handler: ra.Handle,
	}
}

func (ra *RunAgent) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input struct {
		Name   string `param:"name,required"`
		Prompt string `param:"prompt,required"`
	}
	if err := mcpx.MapArguments(req.Params.Arguments, &input); err != nil {
		return mcpx.NewToolResultErrorf("failed to parse arguments", err), nil
	}
	a := ra.GetAgent(input.Name)
	prompt := strings.Join([]string{
		input.Prompt,
		"Note: Only your final response message will be provided back to the user.",
		"This last message should contain all of the relevant information.",
	}, "\n\n")
	if err := a.Run(ctx, prompt, true); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(a.LastMessageJSON()), nil
}
