package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/icholy/sloppy/internal/builtin"
	"github.com/icholy/sloppy/internal/sloppy"
	"github.com/icholy/sloppy/internal/termcolor"
)

func main() {
	var prompt string
	var configPath string
	var useBuiltin bool
	var useV2ApplyDiff bool
	flag.StringVar(&configPath, "config", "", "configuration file")
	flag.BoolVar(&useBuiltin, "builtin", true, "use built-in tools")
	flag.StringVar(&prompt, "prompt", "", "use this prompt and then exit")
	flag.BoolVar(&useV2ApplyDiff, "apply_diff.v2", false, "use v2 of apply_diff tool")
	flag.Parse()
	var opt sloppy.AgentOptions
	ctx := context.Background()
	if configPath != "" {
		config, err := ReadConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}
		tools, err := config.ListTools(ctx)
		if err != nil {
			log.Fatal(err)
		}
		opt.Tools = append(opt.Tools, tools...)
	}
	if useBuiltin {
		tools := builtin.Tools("builtin",
			&builtin.RunAgent{Options: &opt},
			&builtin.RunCommand{},
			&builtin.ApplyDiff{Threshold: 0.9, V2: useV2ApplyDiff},
			&builtin.WriteFile{},
			&builtin.ReadFile{},
		)
		opt.Tools = append(opt.Tools, tools...)
	}
	agent := sloppy.New(opt)
	if prompt != "" {
		input := &sloppy.RunInput{Prompt: prompt}
		if _, err := agent.Run(ctx, input); err != nil {
			log.Fatal(err)
		}
		return
	}
	fmt.Println("Tell sloppy what to do")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s: ", termcolor.Text("You", termcolor.Blue))
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		switch strings.TrimSpace(text) {
		case "/tools":
			for i, t := range opt.Tools {
				if i > 0 {
					fmt.Println()
				}
				data, _ := json.MarshalIndent(t.Tool.InputSchema, "", "  ")
				fmt.Printf("Tool: %s\nDescription: %s\nSchema: %s\n",
					t.Name,
					t.Tool.Description,
					data,
				)
			}
			continue
		}
		input := &sloppy.RunInput{Prompt: text}
		if _, err := agent.Run(ctx, input); err != nil {
			log.Fatal(err)
		}
	}
}
