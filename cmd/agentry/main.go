package main

import (
	"flag"
	"fmt"
	"os"
	"context"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	mode := flag.String("mode", "dev", "dev|serve|eval")
	conf := flag.String("config", "", "path to .agentry.yaml")
	flag.Parse()

	switch *mode {
	case "dev":
		// start REPL (omitted for brevity)
	case "serve":
		if *conf == "" {
			fmt.Println("need --config")
			os.Exit(1)
		}
		cfg, err := config.Load(*conf)
		if err != nil {
			panic(err)
		}
		reg := tool.Registry{}
		// Register inline Go echo tool
		reg["echo"] = tool.New("echo", "Repeats the input string", func(ctx context.Context, args map[string]any) (string, error) {
			input, ok := args["text"].(string)
			if !ok {
				return "", fmt.Errorf("missing or invalid 'text' arg")
			}
			return input, nil
		})
		// Register other tools from manifest, skipping echo if present
		for _, m := range cfg.Tools {
			if m.Name == "echo" {
				continue
			}
			tl, _ := tool.FromManifest(m)
			reg[m.Name] = tl
		}
		r := router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
		ag := core.New(r, reg, memory.NewInMemory(), nil)
		agents := map[string]*core.Agent{"default": ag}
		server.Serve(agents)
	case "eval":
		// load suite and run eval on default agent
	default:
		fmt.Println("unknown mode")
	}
}
