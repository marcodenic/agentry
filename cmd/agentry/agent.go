//go:build !tools
// +build !tools

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/audit"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

// loadPrimaryAgentRole loads the agent_0 role configuration
func loadPrimaryAgentRole() (*team.RoleConfig, error) {
	// Use the same search paths as core.GetDefaultPrompt but for full role config
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

	for _, path := range candidates {
		if role, err := team.LoadRoleFromFile(path); err == nil {
			debug.Printf("Loaded primary agent role from %s: model=%v", path, role.Model)
			return role, nil
		}
	}

	return nil, fmt.Errorf("agent_0.yaml not found in any search path")
}

// buildAgent constructs an Agent from configuration.
func buildAgent(cfg *config.File) (*core.Agent, error) {
	tool.SetPermissions(cfg.Permissions.Tools)
	// Sandboxing completely removed
	reg := tool.Registry{}
	for _, m := range cfg.Tools {
		tl, err := tool.FromManifest(m)
		if err != nil {
			if errors.Is(err, tool.ErrUnknownBuiltin) {
				debug.Printf("skipping builtin %s: not available", m.Name)
				continue
			}
			return nil, err
		}
		reg[m.Name] = tl
	}

	// Agent delegation tool is registered by team.RegisterAgentTool at runtime.
	var logWriter *audit.Log
	if path := os.Getenv("AGENTRY_AUDIT_LOG"); path != "" {
		if lw, err := audit.Open(path, 1<<20); err == nil {
			logWriter = lw
			reg = tool.WrapWithAudit(reg, lw)
		}
	}

	// Use the first configured model, or mock if none configured
	var client model.Client
	var modelName string

	// Try to load the primary agent role configuration for model override
	primaryRole, roleErr := loadPrimaryAgentRole()
	var roleModel *config.ModelManifest
	if roleErr == nil && primaryRole.Model != nil {
		roleModel = primaryRole.Model
		debug.Printf("Primary agent will use role-specific model: %s/%s", roleModel.Provider, roleModel.Options["model"])
	}

	if roleModel != nil {
		// Use role-specific model configuration
		c, err := model.FromManifest(*roleModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create role-specific model client: %w", err)
		}
		client = c

		// Construct the model name: provider/model
		if roleModel.Options != nil && roleModel.Options["model"] != "" {
			modelName = fmt.Sprintf("%s/%s", roleModel.Provider, roleModel.Options["model"])
		} else {
			modelName = roleModel.Provider
		}
		debug.Printf("Using role-specific model: %s", modelName)
	} else if len(cfg.Models) > 0 {
		// Fall back to global model configuration
		primaryModel := cfg.Models[0]
		c, err := model.FromManifest(primaryModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create primary model client: %w", err)
		}
		client = c

		// Construct the model name: provider/model (e.g., "openai/gpt-4.1-nano")
		if primaryModel.Options != nil && primaryModel.Options["model"] != "" {
			modelName = fmt.Sprintf("%s/%s", primaryModel.Provider, primaryModel.Options["model"])
		} else {
			modelName = primaryModel.Name // fallback to name if no model option
		}
		debug.Printf("Using global model: %s", modelName)
	} else {
		// Fallback to mock if no models configured
		client = model.NewMock()
		modelName = "mock"
		debug.Printf("Using mock model")
	}

	var vec memory.VectorStore
	switch cfg.Vector.Type {
	case "qdrant":
		vec = memory.NewQdrant(cfg.Vector.URL, cfg.Vector.Collection)
	case "faiss":
		vec = memory.NewFaiss(cfg.Vector.URL)
	default:
		vec = memory.NewInMemoryVector()
	}

	ag := core.New(client, modelName, reg, memory.NewInMemory(), vec, nil)

	// Configure error handling for resilience
	ag.ErrorHandling.TreatErrorsAsResults = true
	ag.ErrorHandling.MaxErrorRetries = 3
	ag.ErrorHandling.IncludeErrorContext = true

	// Debug: check what tools the agent actually gets (only in non-TUI mode)
	if os.Getenv("AGENTRY_TUI_MODE") != "1" {
		debug.Printf("buildAgent: registry has %d tools, agent has %d tools", len(reg), len(ag.Tools))
	}

	if logWriter != nil {
		ag.Tracer = trace.NewJSONL(logWriter)
	}
	// No iteration cap

	// Use prompt from role configuration if available, otherwise fall back to GetDefaultPrompt
	var prompt string
	if roleErr == nil && primaryRole != nil && strings.TrimSpace(primaryRole.Prompt) != "" {
		prompt = primaryRole.Prompt
		debug.Printf("Using role-specific prompt from agent_0.yaml")
	} else {
		// Fallback to the existing GetDefaultPrompt logic
		prompt = core.GetDefaultPrompt()
		debug.Printf("Using prompt from GetDefaultPrompt fallback")
	}

	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("no default prompt found: place agent_0.yaml under one of: $AGENTRY_DEFAULT_PROMPT, $AGENTRY_CONFIG_HOME/roles/, ~/.config/agentry/roles/, <exedir>/templates/roles/, ./templates/roles/")
	}

	ag.Prompt = prompt

	// Initialize/override cost manager budgets from config when provided.
	// core.New() already set budgets from env; honor config if specified.
	if cfg.Budget.Tokens > 0 || cfg.Budget.Dollars > 0 {
		ag.Cost = cost.New(cfg.Budget.Tokens, cfg.Budget.Dollars)
	}

	return ag, nil
}
