# SLOPPY

> A sloppy CLI agent which supports MCP servers.

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

Sloppy comes with 2 built-in tools: `edit_file` and `run_command`.
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
