package main

import (
	"context"
	"flag"
	"log"

	"github.com/icholy/sloppy/internal/builtin"
	"github.com/icholy/sloppy/internal/sloppy"
)

func main() {
	var configPath string
	var useBuiltin bool
	flag.StringVar(&configPath, "config", "", "configuration file")
	flag.BoolVar(&useBuiltin, "builtin", true, "use built-in tools")
	flag.Parse()
	var opt sloppy.Options
	if useBuiltin {
		opt.Tools = append(opt.Tools, builtin.Tools()...)
	}
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
	agent := sloppy.New(opt)
	if err := agent.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
