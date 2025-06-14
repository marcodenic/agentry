package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/eval"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	env.Load()
	mode := flag.String("mode", "dev", "dev|serve|eval")
	conf := flag.String("config", "", "path to .agentry.yaml")
	useReal := flag.Bool("use-real", false, "use real OpenAI model")
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
		if *conf == "" {
			fmt.Println("need --config")
			os.Exit(1)
		}
		cfg, err := config.Load(*conf)
		if err != nil {
			panic(err)
		}
		reg := tool.Registry{}
		for _, m := range cfg.Tools {
			tl, _ := tool.FromManifest(m)
			reg[m.Name] = tl
		}
		var client model.Client
		key := os.Getenv("OPENAI_KEY")
		if *useReal || key != "" {
			if key == "" {
				fmt.Println("OPENAI_KEY not set, falling back to mock")
				client = model.NewMock()
			} else {
				fmt.Println("Using real OpenAI model")
				client = model.NewOpenAI(key)
			}
		} else {
			client = model.NewMock()
		}
		r := router.Rules{{IfContains: []string{""}, Client: client}}
		ag := core.New(r, reg, memory.NewInMemory(), nil)
		eval.Run(nil, ag, "tests/eval_suite.json")
	default:
		fmt.Println("unknown mode")
	}
}
