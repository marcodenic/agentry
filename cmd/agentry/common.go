package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/config"
)

type commonOpts struct {
	configPath   string
	theme        string
	keybindsPath string
	credsPath    string
	mcpFlag      string
	saveID       string
	resumeID     string
	ckptID       string
	port         string
	debug        bool
}

func parseCommon(name string, args []string) (*commonOpts, []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	opts := &commonOpts{}
	fs.StringVar(&opts.configPath, "config", "", "path to .agentry.yaml")
	fs.StringVar(&opts.theme, "theme", "", "theme name override")
	fs.StringVar(&opts.keybindsPath, "keybinds", "", "path to keybinds json")
	fs.StringVar(&opts.credsPath, "creds", "", "path to credentials json")
	fs.StringVar(&opts.mcpFlag, "mcp", "", "comma-separated MCP servers")
	fs.StringVar(&opts.saveID, "save-id", "", "save conversation state to this ID")
	fs.StringVar(&opts.resumeID, "resume-id", "", "load conversation state from this ID")
	fs.StringVar(&opts.ckptID, "checkpoint-id", "", "checkpoint session id")
	fs.StringVar(&opts.port, "port", "", "HTTP server port")
	fs.BoolVar(&opts.debug, "debug", false, "enable debug output")
	// max-iter removed: agents run until completion
	_ = fs.Parse(args)

	// Only set config path from first non-flag argument for commands that expect a config file
	if opts.configPath == "" && name == "tui" {
		if fs.NArg() > 0 {
			opts.configPath = fs.Arg(0)
		} else {
			// Look for .agentry.yaml in current directory first
			if _, err := os.Stat(".agentry.yaml"); err == nil {
				opts.configPath = ".agentry.yaml"
			} else {
				// Fall back to config next to executable
				if exe, err := os.Executable(); err == nil {
					if exeDir := filepath.Dir(exe); exeDir != "" {
						executableConfig := filepath.Join(exeDir, ".agentry.yaml")
						if _, err := os.Stat(executableConfig); err == nil {
							opts.configPath = executableConfig
						} else {
							opts.configPath = ".agentry.yaml" // Default fallback
						}
					}
				} else {
					opts.configPath = ".agentry.yaml" // Default fallback
				}
			}
		}
	} else if opts.configPath == "" {
		// For non-TUI commands, use default config resolution without assuming first arg is config
		// Look for .agentry.yaml in current directory first
		if _, err := os.Stat(".agentry.yaml"); err == nil {
			opts.configPath = ".agentry.yaml"
		} else {
			// Fall back to config next to executable
			if exe, err := os.Executable(); err == nil {
				if exeDir := filepath.Dir(exe); exeDir != "" {
					executableConfig := filepath.Join(exeDir, ".agentry.yaml")
					if _, err := os.Stat(executableConfig); err == nil {
						opts.configPath = executableConfig
					} else {
						opts.configPath = ".agentry.yaml" // Default fallback
					}
				}
			} else {
				opts.configPath = ".agentry.yaml" // Default fallback
			}
		}
	}
	return opts, fs.Args()
}

func applyOverrides(cfg *config.File, o *commonOpts) {
	// Handle debug flag by setting environment variable
	if o.debug {
		os.Setenv("AGENTRY_DEBUG", "1")
	}

	if o.theme != "" {
		if cfg.Themes == nil {
			cfg.Themes = map[string]string{}
		}
		cfg.Themes["active"] = o.theme
		cfg.Theme = o.theme
	}
	if cfg.Theme != "" {
		os.Setenv("AGENTRY_THEME", cfg.Theme)
	}
	if o.keybindsPath != "" {
		if b, err := os.ReadFile(o.keybindsPath); err == nil {
			_ = json.Unmarshal(b, &cfg.Keybinds)
		}
	}
	if o.credsPath != "" {
		if b, err := os.ReadFile(o.credsPath); err == nil {
			_ = json.Unmarshal(b, &cfg.Credentials)
		}
	}
	if o.mcpFlag != "" {
		if cfg.MCPServers == nil {
			cfg.MCPServers = map[string]string{}
		}
		parts := strings.Split(o.mcpFlag, ",")
		for i, p := range parts {
			cfg.MCPServers[fmt.Sprintf("srv%d", i+1)] = strings.TrimSpace(p)
		}
	}
}

// emitJSON outputs a JSON response to stdout
func emitJSON(data any) {
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Printf(`{"ok": false, "error": "json marshal failed: %v"}`, err)
		return
	}
	fmt.Println(string(b))
}

// osBackgroundContext provides a cancellable background context.
func osBackgroundContext() context.Context { return context.Background() }
