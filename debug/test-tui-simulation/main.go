package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func main() {
	// Enable debug logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[DEBUG] Simulating exact TUI scenario")

	// Show working directory
	if cwd, err := os.Getwd(); err == nil {
		fmt.Printf("🏠 Working directory: %s\n", cwd)
	}

	// Load config exactly like TUI
	cfg, err := config.Load(".agentry.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build agent exactly like TUI
	agent0, err := buildAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to build agent: %v", err)
	}

	fmt.Printf("🤖 Agent 0 created with %d tools\n", len(agent0.Tools))

	// Create team context exactly like TUI
	teamCtx, err := team.NewTeam(agent0, 10, "test")
	if err != nil {
		log.Fatalf("Failed to create team context: %v", err)
	}

	fmt.Printf("🎯 Team context created\n")

	// Set up context with team exactly like TUI does
	ctx := team.WithContext(context.Background(), teamCtx)

	// Test the exact input that fails in TUI: "coder read TODO.md"
	input := "coder read TODO.md"
	fmt.Printf("\n📝 Running Agent 0 with input: %s\n", input)
	fmt.Println("==========================================")

	result, err := agent0.Run(ctx, input)
	if err != nil {
		fmt.Printf("❌ Agent 0 failed: %v\n", err)
	} else {
		fmt.Printf("✅ Agent 0 succeeded!\n")
		fmt.Printf("📄 Result: %s\n", result)
	}

	fmt.Println("==========================================")
	fmt.Printf("🔍 This simulates exactly what should happen in TUI mode\n")
}

// buildAgent constructs an Agent from configuration - copied from agent.go
func buildAgent(cfg *config.File) (*core.Agent, error) {
	tool.SetPermissions(cfg.Permissions.Tools)
	tool.SetSandboxEngine(cfg.Sandbox.Engine)
	reg := tool.Registry{}
	for _, m := range cfg.Tools {
		tl, err := tool.FromManifest(m)
		if err != nil {
			if err == tool.ErrUnknownBuiltin {
				fmt.Printf("skipping builtin %s: not available\n", m.Name)
				continue
			}
			return nil, err
		}
		reg[m.Name] = tl
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

	if cfg.MaxIterations > 0 {
		ag.MaxIterations = cfg.MaxIterations
	}

	// Use default prompt for main agent - team.go will load role configs when spawning
	ag.Prompt = "You are Agent 0, the system orchestrator. You can delegate to specialized agents using the agent tool."

	return ag, nil
}
