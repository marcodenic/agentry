package core

import (
	"os"
	"path/filepath"
	"github.com/marcodenic/agentry/internal/debug"
	"gopkg.in/yaml.v3"
)

// RoleConfig represents a role configuration (duplicated to avoid import cycle)
type roleConfig struct {
	Name   string `yaml:"name"`
	Prompt string `yaml:"prompt"`
}

// GetDefaultPrompt loads the canonical agent_0 prompt from the role file
func GetDefaultPrompt() string {
	// Resolve candidate search paths in priority order
	// 1) Explicit env override
	// 2) XDG config (~/.config/agentry/roles/agent_0.yaml or AGENTRY_CONFIG_HOME)
	// 3) Next to executable (bin/templates/roles/agent_0.yaml)
	// 4) Current working dir (templates/roles/agent_0.yaml)

	candidates := make([]string, 0, 4)

	if p := os.Getenv("AGENTRY_DEFAULT_PROMPT"); p != "" {
		candidates = append(candidates, p)
	}

	// XDG config
	cfgHome := os.Getenv("AGENTRY_CONFIG_HOME")
	if cfgHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfgHome = filepath.Join(home, ".config", "agentry")
		}
	}
	if cfgHome != "" {
		candidates = append(candidates, filepath.Join(cfgHome, "roles", "agent_0.yaml"))
	}

	// Executable dir
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, "templates", "roles", "agent_0.yaml"))
	}

	// Working dir
	candidates = append(candidates, filepath.Join("templates", "roles", "agent_0.yaml"))

	var b []byte
	for _, p := range candidates {
		if data, e := os.ReadFile(p); e == nil {
			b = data
			break
		}
	}
	if b == nil {
		debug.Printf("Default prompt file not found in any search path; using minimal embedded fallback.")
		// Provide a minimal safe fallback so agent remains functional.
		return "You are Agentry, a helpful, tool-using assistant. Always be concise, truthful, and cite any tool usage when relevant."
	}

	var role roleConfig
	if err := yaml.Unmarshal(b, &role); err != nil {
		return ""
	}

	return role.Prompt
}

func defaultPrompt() string { return GetDefaultPrompt() }

// No code-embedded prompt beyond a minimal fallback above; prefer user-editable files.
