package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/eval"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/tui"
)

func main() {
	env.Load()
	mode := flag.String("mode", "dev", "dev|serve|eval|tui")
	conf := flag.String("config", "", "path to .agentry.yaml")
	flag.Parse()

	switch *mode {
	case "dev":
		cfg, err := config.Load("examples/.agentry.yaml")
		if err != nil {
			panic(err)
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}

		// tiny REPL
		sc := bufio.NewScanner(os.Stdin)
		fmt.Println("Agentry REPL – Ctrl-D to quit")
		for {
			fmt.Print("> ")
			if !sc.Scan() {
				break
			}
			line := sc.Text()
			out, err := ag.Run(context.Background(), line)
			if err != nil {
				fmt.Println("ERR:", err)
				continue
			}
			fmt.Println(out)
		}
	case "serve":
		if *conf == "" {
			fmt.Println("need --config")
			os.Exit(1)
		}
		cfg, err := config.Load(*conf)
		if err != nil {
			panic(err)
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}
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
		key := os.Getenv("OPENAI_KEY")
		if key != "" {
			for i, m := range cfg.Models {
				if m.Name == "openai" {
					if m.Options == nil {
						m.Options = map[string]string{}
					}
					cfg.Models[i].Options["key"] = key
				}
			}
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}
		suite := "tests/eval_suite.json"
		if key != "" {
			suite = "tests/openai_eval_suite.json"
		}
		eval.Run(nil, ag, suite)
	case "tui":
		if *conf == "" {
			fmt.Println("need --config")
			os.Exit(1)
		}
		cfg, err := config.Load(*conf)
		if err != nil {
			panic(err)
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}
		p := tea.NewProgram(tui.New(ag))
		if err := p.Start(); err != nil {
			panic(err)
		}
	default:
		fmt.Println("unknown mode")
	}
}
