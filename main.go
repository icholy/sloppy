package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/icholy/sloppy/internal/builtin"
	"github.com/icholy/sloppy/internal/sloppy"
	"github.com/icholy/sloppy/internal/termcolor"
)

func main() {
	var prompt string
	var configPath string
	var useBuiltin bool
	flag.StringVar(&configPath, "config", "", "configuration file")
	flag.BoolVar(&useBuiltin, "builtin", true, "use built-in tools")
	flag.StringVar(&prompt, "prompt", "", "use this prompt and then exit")
	flag.Parse()
	var opt sloppy.Options
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
		tools := builtin.Tools("Sloppy Built-in tools",
			&builtin.RunAgent{Options: &opt},
			&builtin.EditFile{},
			&builtin.RunCommand{},
		)
		opt.Tools = append(opt.Tools, tools...)
	}
	agent := sloppy.New(opt)
	if prompt != "" {
		if err := agent.Run(ctx, prompt, true); err != nil {
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
		if err := agent.Run(ctx, scanner.Text(), true); err != nil {
			log.Fatal(err)
		}
	}
}
