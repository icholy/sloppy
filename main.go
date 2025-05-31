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
	var driver sloppy.Driver
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
		driver.Tools = append(driver.Tools, tools...)
	}
	if useBuiltin {
		tools := builtin.Tools("builtin",
			// &builtin.RunAgent{Options: &opt},
			&builtin.RunCommand{},
			&builtin.ApplyDiff{Threshold: 0.9, V2: useV2ApplyDiff},
			&builtin.WriteFile{},
			&builtin.ReadFile{},
		)
		driver.Tools = append(driver.Tools, tools...)
	}
	driver.Agent = sloppy.NewAnthropicAgent(nil)
	if prompt != "" {
		if err := driver.Loop(ctx, prompt); err != nil {
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
			for i, t := range driver.Tools {
				if i > 0 {
					fmt.Println()
				}
				data, _ := json.MarshalIndent(t.Tool.InputSchema, "", "  ")
				fmt.Printf("Tool: %s\nDescription: %s\nSchema: %s\n",
					t.Alias,
					t.Tool.Description,
					data,
				)
			}
			continue
		}
		if err := driver.Loop(ctx, text); err != nil {
			log.Fatal(err)
		}
	}
}
