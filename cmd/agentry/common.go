package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/model"
)

type commonOpts struct {
	configPath     string
	theme          string
	keybindsPath   string
	credsPath      string
	mcpFlag        string
	saveID         string
	resumeID       string
	ckptID         string
	port           string
	debug          bool
	disableTools   bool
	allowTools     string
	denyTools      string
	disableContext bool
	auditLog       string

	// New flags (prefer flags over env vars)
	maxIter     int // 0 = unlimited
	httpTimeout int // seconds
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
	fs.BoolVar(&opts.disableTools, "disable-tools", false, "disable tool filtering entirely")
	fs.StringVar(&opts.allowTools, "allow-tools", "", "comma-separated list of additional tools to include")
	fs.StringVar(&opts.denyTools, "deny-tools", "", "comma-separated list of tools to exclude")
	fs.BoolVar(&opts.disableContext, "disable-context", false, "disable context pipeline")
	fs.StringVar(&opts.auditLog, "audit-log", "", "path to audit log file")
	// Debug/diagnostic flags
	fs.IntVar(&opts.maxIter, "max_iter", 0, "limit agent iterations (0=unlimited)")
	fs.IntVar(&opts.maxIter, "max-iter", 0, "limit agent iterations (0=unlimited)")
	fs.IntVar(&opts.httpTimeout, "http_timeout", 300, "HTTP client timeout in seconds")
	fs.IntVar(&opts.httpTimeout, "http-timeout", 300, "HTTP client timeout in seconds")
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
	// Handle debug flag by enabling debug output dynamically
	if o.debug {
		debug.EnableDebug()
	}

	// Handle tool filtering flags by modifying config directly
	if o.disableTools {
		// Clear tool permissions to allow all tools
		cfg.Permissions.Tools = nil
	}
	if o.allowTools != "" {
		// Replace tools list with only the specified tools (restrictive)
		allowList := strings.Split(o.allowTools, ",")
		allowSet := make(map[string]bool)
		for _, tool := range allowList {
			allowSet[strings.TrimSpace(tool)] = true
		}

		// Filter tools to only include allowed ones
		var filteredTools []config.ToolManifest
		for _, tool := range cfg.Tools {
			if allowSet[tool.Name] {
				filteredTools = append(filteredTools, tool)
			}
		}
		cfg.Tools = filteredTools

		// Also set permissions to the allow list
		cfg.Permissions.Tools = allowList
	}
	if o.denyTools != "" {
		// Remove specified tools from config
		denyList := strings.Split(o.denyTools, ",")
		denySet := make(map[string]bool)
		for _, tool := range denyList {
			denySet[strings.TrimSpace(tool)] = true
		}

		// Filter out denied tools from the tools list
		var filteredTools []config.ToolManifest
		for _, tool := range cfg.Tools {
			if !denySet[tool.Name] {
				filteredTools = append(filteredTools, tool)
			}
		}
		cfg.Tools = filteredTools

		// Also remove from permissions if present
		if cfg.Permissions.Tools != nil {
			var filteredPerms []string
			for _, tool := range cfg.Permissions.Tools {
				if !denySet[tool] {
					filteredPerms = append(filteredPerms, tool)
				}
			}
			cfg.Permissions.Tools = filteredPerms
		}
	}

	// Handle context and audit flags
	if o.disableContext {
		os.Setenv("AGENTRY_DISABLE_CONTEXT", "1")
	}
	if o.auditLog != "" {
		os.Setenv("AGENTRY_AUDIT_LOG", o.auditLog)
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

	// Apply model HTTP timeout (flags take precedence over env)
	if o.httpTimeout > 0 {
		model.SetHTTPTimeout(o.httpTimeout)
	}
}
