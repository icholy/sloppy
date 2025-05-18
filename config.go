package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/icholy/sloppy/internal/sloppy"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type Config struct {
	MCPServers map[string]*MCPServerConfig `json:"mcpServers"`
}

func ReadConfig(name string) (*Config, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", name, err)
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", name, err)
	}
	return &config, nil
}

func (c *Config) ListTools(ctx context.Context) ([]sloppy.Tool, error) {
	var tools []sloppy.Tool
	for name, opts := range c.MCPServers {
		c, err := client.NewStdioMCPClient(opts.Command, nil, opts.Args...)
		if err != nil {
			return nil, fmt.Errorf("failed to create mcp client: %s: %w", name, err)
		}
		if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
			return nil, fmt.Errorf("failed to initialize mcp client: %s: %w", name, err)
		}
		tt, err := sloppy.ListClientTools(ctx, name, c)
		if err != nil {
			return nil, fmt.Errorf("failed to list mcp tools: %s: %w", name, err)
		}
		tools = append(tools, tt...)
	}
	return tools, nil
}
