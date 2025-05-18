# SLOPPY

> A sloppy CLI agent which supports MCP servers.


## Agent Delegation

Sloppy's main agent can delegate subtasks by spawning child agents, which are themselves new, independent instances of Sloppy. These child agents execute their assigned portion of a task and return only a brief summary of their results (not their full output or detailed execution logs). This prevents the main agent's context window from being filled with unnecessary detail and helps the parent agent stay focused on the overall task.

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

Sloppy comes with 5 built-in tools:

- `run_command`: Executes shell commands
- `run_agent`: Delegates subtasks to child agents
- `apply_diff`: Applies search/replace changes to a text file using diff blocks
- `read_file`: Reads content from a file, optionally specifying line ranges
- `write_file`: Creates or replaces a file with specified content

**Note**: These can be disabled using the `--builtin` flag.

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
