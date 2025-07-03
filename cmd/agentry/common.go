package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/tool"
	"gopkg.in/yaml.v3"
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
	maxIter      int
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
	fs.IntVar(&opts.maxIter, "max-iter", 0, "max iterations per run")
	_ = fs.Parse(args)
	if opts.configPath == "" {
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
	}
	return opts, fs.Args()
}

func applyOverrides(cfg *config.File, o *commonOpts) {
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

// applyAgent0RoleConfig applies the agent_0.yaml role configuration to restrict the system agent's tools
func applyAgent0RoleConfig(agent *core.Agent) error {
	// Find the templates/roles directory
	roleDir := findRoleTemplatesDir()
	if roleDir == "" {
		return fmt.Errorf("templates/roles directory not found")
	}

	// Load agent_0.yaml configuration
	roleFile := filepath.Join(roleDir, "agent_0.yaml")
	data, err := os.ReadFile(roleFile)
	if err != nil {
		return fmt.Errorf("failed to read agent_0.yaml: %v", err)
	}

	// Parse YAML configuration
	var config struct {
		Name     string   `yaml:"name"`
		Prompt   string   `yaml:"prompt"`
		Builtins []string `yaml:"builtins,omitempty"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse agent_0.yaml: %v", err)
	}

	// Apply tool restrictions if builtins are specified
	if len(config.Builtins) > 0 {
		filteredTools := make(tool.Registry)

		for _, toolName := range config.Builtins {
			if existingTool, ok := agent.Tools[toolName]; ok {
				filteredTools[toolName] = existingTool
			}
		}

		agent.Tools = filteredTools
	}

	// Apply the agent_0 prompt
	if config.Prompt != "" {
		agent.Prompt = config.Prompt
	}

	// Inject dynamic agent status after prompt is set
	injectAgentStatus(agent, nil)

	return nil
}

// injectAgentStatus dynamically injects current agent status into Agent 0's prompt
func injectAgentStatus(agent *core.Agent, teamCaller interface{}) {
	// Get available agents - we know these exist based on our configuration
	availableAgents := []string{
		"coder", "tester", "writer", "devops", "designer",
		"deployer", "editor", "reviewer", "researcher", "team_planner",
	}

	// Build a concise agent status section for the prompt
	statusSection := "\n\n## 🤖 CURRENT AGENT STATUS\n\n"
	statusSection += "**Available Agents:** "
	for i, agentName := range availableAgents {
		if i > 0 {
			statusSection += ", "
		}
		statusSection += agentName
	}
	statusSection += "\n\n**All agents are ready for immediate delegation via the `agent` tool.**\n\n"

	// Inject the status into the prompt if it doesn't already contain it
	if !strings.Contains(agent.Prompt, "CURRENT AGENT STATUS") {
		agent.Prompt = agent.Prompt + statusSection
	}
}

// findRoleTemplatesDir searches for the templates/roles directory
func findRoleTemplatesDir() string {
	// Try current directory first
	if _, err := os.Stat("templates/roles"); err == nil {
		return "templates/roles"
	}

	// Try walking up the directory tree
	cwd, _ := os.Getwd()
	dir := cwd
	for {
		templatePath := filepath.Join(dir, "templates", "roles")
		if _, err := os.Stat(templatePath); err == nil {
			return templatePath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return ""
}
