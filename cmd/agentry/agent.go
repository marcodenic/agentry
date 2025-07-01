package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/audit"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/pkg/memstore"
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
				fmt.Printf("skipping builtin %s: not available\n", m.Name)
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
				fmt.Printf("❌ Agent tool: no team found in context\n")
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

	var rules router.Rules
	for _, rr := range cfg.Routes {
		c, ok := clients[rr.Model]
		if !ok {
			return nil, fmt.Errorf("model %s not found", rr.Model)
		}
		rules = append(rules, router.Rule{Name: rr.Model, IfContains: rr.IfContains, Client: c})
	}
	if len(rules) == 0 {
		rules = router.Rules{{Name: "mock", IfContains: []string{""}, Client: model.NewMock()}}
	}

	var store memstore.KV
	memURI := cfg.Memory
	if memURI == "" {
		memURI = cfg.Store
	}
	if memURI == "" {
		memURI = "mem"
	}
	if memURI != "" {
		s, err := memstore.StoreFactory(memURI)
		if err != nil {
			return nil, err
		}
		store = s
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

	ag := core.New(rules, reg, memory.NewInMemory(), store, vec, nil)
	
	// Debug: check what tools the agent actually gets
	fmt.Printf("🔧 buildAgent: registry has %d tools, agent has %d tools\n", len(reg), len(ag.Tools))
	
	if logWriter != nil {
		ag.Tracer = trace.NewJSONL(logWriter)
	}
	if cfg.MaxIterations > 0 {
		ag.MaxIterations = cfg.MaxIterations
	}
	
	// Use default prompt for main agent - team.go will load role configs when spawning
	ag.Prompt = "You are Agent 0, the system orchestrator. You can delegate to specialized agents using the agent tool."
	
	// Initialize cost manager for token/cost tracking
	ag.Cost = cost.New(0, 0.0) // No budget limits, just tracking
	
	return ag, nil
}

func runCostCmd(args []string)   { fmt.Println("Cost command not implemented yet") }
func runPProfCmd(args []string)  { fmt.Println("PProf command not implemented yet") }
func runPluginCmd(args []string) { fmt.Println("Plugin command not implemented yet") }
func runToolCmd(args []string)   { fmt.Println("Tool command not implemented yet") }
