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

		clients := map[string]model.Client{}
		for _, m := range cfg.Models {
			c, err := model.FromManifest(m)
			if err != nil {
				panic(err)
			}
			clients[m.Name] = c
		}

		var rules router.Rules
		for _, rr := range cfg.Routes {
			c, ok := clients[rr.Model]
			if !ok {
				panic(fmt.Errorf("model %s not found", rr.Model))
			}
			rules = append(rules, router.Rule{IfContains: rr.IfContains, Client: c})
		}
		if len(rules) == 0 {
			rules = router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
		}
		ag := core.New(rules, reg, memory.NewInMemory(), nil)
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

		clients := map[string]model.Client{}
		for _, m := range cfg.Models {
			c, err := model.FromManifest(m)
			if err != nil {
				panic(err)
			}
			clients[m.Name] = c
		}

		var (
			suite = "tests/eval_suite.json"
			key   = os.Getenv("OPENAI_KEY")
		)
		if key != "" {
			if c, ok := clients["openai"]; ok {
				clients["openai"] = model.NewOpenAI(key)
				suite = "tests/openai_eval_suite.json"
				_ = c
			}
		}

		var rules router.Rules
		for _, rr := range cfg.Routes {
			c, ok := clients[rr.Model]
			if !ok {
				panic(fmt.Errorf("model %s not found", rr.Model))
			}
			rules = append(rules, router.Rule{IfContains: rr.IfContains, Client: c})
		}
		if len(rules) == 0 {
			rules = router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
		}
		ag := core.New(rules, reg, memory.NewInMemory(), nil)
		eval.Run(nil, ag, suite)
	default:
		fmt.Println("unknown mode")
	}
}
