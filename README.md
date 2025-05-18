# SLOPPY

> A sloppy CLI agent which supports MCP servers.


## Agent Delegation

Sloppy can delegate parts of a task to child agents. When a task is complex or has distinct steps, the main agent can spawn child agents to handle specific subtasks.

To keep things efficient, child agents don’t add their full output or execution details to the main agent’s context. Instead, they return only a brief summary of their results. This helps the main agent stay focused and prevents the context from growing too large, making it easier to manage complex workflows.

### Install

```
go install github.com/icholy/sloppy@latest
```

### Run

```
$ sloppy
Tell sloppy what to do
You: I'd like some slop.
```

### Tools

Sloppy comes with 2 built-in tools: `edit_file`, `run_command`, and `run_agent`.
These can be disabled using the `--builtin` flag.

### MCP

Additional MCP server may be configured using a `sloppy.json` file.
The file shares the same configuration format as Cursor's MCP configs.

```json
{
  "mcpServers": {
    "server-name": {
      "command": "semgrep-cli-mcp",
      "args": ["--configs", "my-semgrep-configs"]
    }
  }
}
```

Use the `--config` flag to load the configuration file.

```
sloppy --config ./sloppy.json
```
