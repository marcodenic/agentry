//go:build !tools
// +build !tools

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
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

// buildAgent constructs an Agent from configuration.
func buildAgent(cfg *config.File) (*core.Agent, error) {
	tool.SetPermissions(cfg.Permissions.Tools)
	tool.SetSandboxEngine(cfg.Sandbox.Engine)
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

	// Replace the default agent tool with proper team-context implementation
	if _, hasAgent := reg["agent"]; hasAgent {
		reg["agent"] = tool.NewWithSchema("agent", "Delegate to another agent", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{"type": "string"},
				"input": map[string]any{"type": "string"},
			},
			"required": []string{"agent", "input"},
			"example": map[string]any{
				"agent": "Agent1",
				"input": "Hello, how are you?",
			},
		}, func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["agent"].(string)
			input, _ := args["input"].(string)
			t, ok := team.FromContext(ctx)
			if !ok || t == nil {
				debug.Printf("Agent tool: no team found in context")
				return "", fmt.Errorf("team not found in context")
			}
			return t.Call(ctx, name, input)
		})
	}
	var logWriter *audit.Log
	if path := os.Getenv("AGENTRY_AUDIT_LOG"); path != "" {
		if lw, err := audit.Open(path, 1<<20); err == nil {
			logWriter = lw
			reg = tool.WrapWithAudit(reg, lw)
		}
	}

	clients := map[string]model.Client{}
	for _, m := range cfg.Models {
		c, err := model.FromManifest(m)
		if err != nil {
			return nil, err
		}
		clients[m.Name] = c
	}

	// Use the first configured model, or mock if none configured
	var client model.Client
	var modelName string

	if len(cfg.Models) > 0 {
		primaryModel := cfg.Models[0]

		c, ok := clients[primaryModel.Name]
		if !ok {
			return nil, fmt.Errorf("primary model %s not found", primaryModel.Name)
		}
		client = c

		// Construct the model name: provider/model (e.g., "openai/gpt-4.1-nano")
		if primaryModel.Options != nil && primaryModel.Options["model"] != "" {
			modelName = fmt.Sprintf("%s/%s", primaryModel.Provider, primaryModel.Options["model"])
		} else {
			modelName = primaryModel.Name // fallback to name if no model option
		}
	} else {
		// Fallback to mock if no models configured
		client = model.NewMock()
		modelName = "mock"
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

	// Resolve default prompt from user-editable files; fail if missing
	ag.Prompt = core.GetDefaultPrompt()
	if strings.TrimSpace(ag.Prompt) == "" {
		return nil, fmt.Errorf("no default prompt found: place agent_0.yaml under one of: $AGENTRY_DEFAULT_PROMPT, $AGENTRY_CONFIG_HOME/roles/, ~/.config/agentry/roles/, <exedir>/templates/roles/, ./templates/roles/")
	}

	// Initialize cost manager for token/cost tracking
	ag.Cost = cost.New(0, 0.0) // No budget limits, just tracking

	return ag, nil
}

// Stub functions for commands that are only available with tools build tag
func runPProfCmd(_ []string) {
	fmt.Println("PProf command not available (build with --tools flag)")
}
