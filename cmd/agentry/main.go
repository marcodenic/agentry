package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/eval"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/internal/tui"
)

func main() {
	env.Load()
	if len(os.Args) < 2 {
		fmt.Println("Usage: agentry [dev|serve|tui|eval] [--config path/to/config.yaml]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	conf := fs.String("config", "", "path to .agentry.yaml")
	theme := fs.String("theme", "", "theme name override")
	keybinds := fs.String("keybinds", "", "path to keybinds json")
	credsPath := fs.String("creds", "", "path to credentials json")
	mcpFlag := fs.String("mcp", "", "comma-separated MCP servers")
	_ = fs.Parse(args)
	var configPath string
	if *conf != "" {
		configPath = *conf
	} else if fs.NArg() > 0 {
		configPath = fs.Arg(0)
	} else {
		configPath = "examples/.agentry.yaml"
	}

	switch cmd {
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
			if strings.HasPrefix(line, "converse") {
				rest := strings.TrimSpace(strings.TrimPrefix(line, "converse"))
				n := 2
				topic := ""
				if rest != "" {
					fields := strings.Fields(rest)
					if len(fields) > 0 {
						if v, err := strconv.Atoi(fields[0]); err == nil && v > 0 {
							n = v
							rest = strings.TrimSpace(rest[len(fields[0]):])
						}
					}
					topic = strings.TrimSpace(rest)
				}
				if topic == "" {
					topic = "Hello agents, let's chat!"
				} else if (strings.HasPrefix(topic, "\"") && strings.HasSuffix(topic, "\"")) ||
					(strings.HasPrefix(topic, "'") && strings.HasSuffix(topic, "'")) {
					topic = strings.Trim(topic, "'\"")
				}
				converse.Repl(ag, n, topic)
				continue
			}
			out, err := ag.Run(context.Background(), line)
			if err != nil {
				fmt.Println("ERR:", err)
				continue
			}
			fmt.Println(out)
		}
	case "serve":
		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			os.Exit(1)
		}
		if *theme != "" {
			if cfg.Themes == nil {
				cfg.Themes = map[string]string{}
			}
			cfg.Themes["active"] = *theme
		}
		if *keybinds != "" {
			if b, err := os.ReadFile(*keybinds); err == nil {
				_ = json.Unmarshal(b, &cfg.Keybinds)
			}
		}
		if *credsPath != "" {
			if b, err := os.ReadFile(*credsPath); err == nil {
				_ = json.Unmarshal(b, &cfg.Credentials)
			}
		}
		if *mcpFlag != "" {
			if cfg.MCPServers == nil {
				cfg.MCPServers = map[string]string{}
			}
			parts := strings.Split(*mcpFlag, ",")
			for i, p := range parts {
				cfg.MCPServers[fmt.Sprintf("srv%d", i+1)] = strings.TrimSpace(p)
			}
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}
		if cfg.Metrics {
			if _, err := trace.Init(); err != nil {
				fmt.Printf("trace init: %v\n", err)
			}
		}
		agents := map[string]*core.Agent{"default": ag}
		fmt.Println("Serving HTTP on :8080")
		server.Serve(agents, cfg.Metrics)
	case "eval":
		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			os.Exit(1)
		}
		if *theme != "" {
			if cfg.Themes == nil {
				cfg.Themes = map[string]string{}
			}
			cfg.Themes["active"] = *theme
		}
		if *keybinds != "" {
			if b, err := os.ReadFile(*keybinds); err == nil {
				_ = json.Unmarshal(b, &cfg.Keybinds)
			}
		}
		if *credsPath != "" {
			if b, err := os.ReadFile(*credsPath); err == nil {
				_ = json.Unmarshal(b, &cfg.Credentials)
			}
		}
		if *mcpFlag != "" {
			if cfg.MCPServers == nil {
				cfg.MCPServers = map[string]string{}
			}
			parts := strings.Split(*mcpFlag, ",")
			for i, p := range parts {
				cfg.MCPServers[fmt.Sprintf("srv%d", i+1)] = strings.TrimSpace(p)
			}
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
		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			os.Exit(1)
		}
		if *theme != "" {
			if cfg.Themes == nil {
				cfg.Themes = map[string]string{}
			}
			cfg.Themes["active"] = *theme
		}
		if *keybinds != "" {
			if b, err := os.ReadFile(*keybinds); err == nil {
				_ = json.Unmarshal(b, &cfg.Keybinds)
			}
		}
		if *credsPath != "" {
			if b, err := os.ReadFile(*credsPath); err == nil {
				_ = json.Unmarshal(b, &cfg.Credentials)
			}
		}
		if *mcpFlag != "" {
			if cfg.MCPServers == nil {
				cfg.MCPServers = map[string]string{}
			}
			parts := strings.Split(*mcpFlag, ",")
			for i, p := range parts {
				cfg.MCPServers[fmt.Sprintf("srv%d", i+1)] = strings.TrimSpace(p)
			}
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}
		p := tea.NewProgram(tui.New(ag))
		if err := p.Start(); err != nil {
			panic(err)
		}
	case "plugin":
		runPluginCmd(args)
	default:
		fmt.Println("unknown command. Usage: agentry [dev|serve|tui|eval] [--config path/to/config.yaml]")
		os.Exit(1)
	}
}
