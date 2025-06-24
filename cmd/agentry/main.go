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
	"time"

	"github.com/google/uuid"

	tea "github.com/charmbracelet/bubbletea"
	agentry "github.com/marcodenic/agentry/internal"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/eval"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/internal/tui"
	"github.com/marcodenic/agentry/pkg/flow"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func main() {
	env.Load()
	if len(os.Args) < 2 {
		fmt.Println("Usage: agentry [dev|serve|tui|eval|flow|version] [--config path/to/config.yaml]")
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
	saveID := fs.String("save-id", "", "save conversation state to this ID")
	resumeID := fs.String("resume-id", "", "load conversation state from this ID")
	ckptID := fs.String("checkpoint-id", "", "checkpoint session id")
	teamSize := fs.Int("team", 0, "number of agents for team chat")
	topic := fs.String("topic", "", "team chat topic")
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
		if cfg.Theme != "" {
			os.Setenv("AGENTRY_THEME", cfg.Theme)
		}
		ag, err := buildAgent(cfg)
		if err != nil {
			panic(err)
		}
		if *resumeID != "" {
			_ = ag.LoadState(context.Background(), *resumeID)
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
			col := trace.NewCollector(nil)
			ag.Tracer = col
			out, err := ag.Run(context.Background(), line)
			if err != nil {
				fmt.Println("ERR:", err)
				continue
			}
			sum := trace.Analyze(line, col.Events())
			fmt.Println(out)
			fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
			if *saveID != "" {
				_ = ag.SaveState(context.Background(), *saveID)
			}
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
			cfg.Theme = *theme
		}
		if cfg.Theme != "" {
			os.Setenv("AGENTRY_THEME", cfg.Theme)
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
		// Session cleanup goroutine
		if dur, err := time.ParseDuration(cfg.SessionTTL); err == nil && dur > 0 {
			if cl, ok := ag.Store.(memstore.Cleaner); ok {
				go func() {
					ticker := time.NewTicker(time.Hour)
					defer ticker.Stop()
					for range ticker.C {
						_ = cl.Cleanup(context.Background(), "history", dur)
					}
				}()
			}
		}
		// Checkpoint logic
		if *ckptID != "" {
			ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(*ckptID))
			_ = ag.Resume(context.Background())
		}
		if *resumeID != "" {
			_ = ag.LoadState(context.Background(), *resumeID)
		}
		if cfg.Metrics {
			if _, err := trace.Init(cfg.Collector); err != nil {
				fmt.Printf("trace init: %v\n", err)
			}
			ag.Tracer = trace.NewOTel()
		}
		agents := map[string]*core.Agent{"default": ag}
		fmt.Println("Serving HTTP on :8080")
		server.Serve(agents, cfg.Metrics, *saveID, *resumeID)
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
			cfg.Theme = *theme
		}
		if cfg.Theme != "" {
			os.Setenv("AGENTRY_THEME", cfg.Theme)
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
		if *ckptID != "" {
			ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(*ckptID))
			_ = ag.Resume(context.Background())
		}
		suite := "tests/eval_suite.json"
		if key != "" {
			suite = "tests/openai_eval_suite.json"
		}
		eval.Run(nil, ag, suite)
		if *saveID != "" {
			_ = ag.SaveState(context.Background(), *saveID)
		}
	case "flow":
		f, err := flow.Load(configPath)
		if err != nil {
			fmt.Printf("failed to load flow: %v\n", err)
			os.Exit(1)
		}
		outs, err := flow.Run(context.Background(), f, tool.DefaultRegistry(), nil)
		if err != nil {
			fmt.Printf("flow error: %v\n", err)
			os.Exit(1)
		}
		for _, o := range outs {
			fmt.Println(o)
		}
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
			cfg.Theme = *theme
		}
		if cfg.Theme != "" {
			os.Setenv("AGENTRY_THEME", cfg.Theme)
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
		if *ckptID != "" {
			ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(*ckptID))
			_ = ag.Resume(context.Background())
		}
		if *resumeID != "" {
			_ = ag.LoadState(context.Background(), *resumeID)
		}
		size := 1
		if *teamSize > 0 {
			size = *teamSize
		}
		cm, err := tui.NewChat(ag, size, *topic)
		if err != nil {
			panic(err)
		}
		p := tea.NewProgram(cm)
		if err := p.Start(); err != nil {
			panic(err)
		}
		if *saveID != "" {
			_ = ag.SaveState(context.Background(), *saveID)
		}
	case "plugin":
		runPluginCmd(args)
	case "tool":
		runToolCmd(args)
	case "version":
		fmt.Printf("agentry %s\n", agentry.Version)
	default:
		fmt.Println("unknown command. Usage: agentry [dev|serve|tui|eval|flow|version] [--config path/to/config.yaml]")
		os.Exit(1)
	}
}
