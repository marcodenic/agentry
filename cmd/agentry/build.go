package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
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
	if path := os.Getenv("AGENTRY_AUDIT_LOG"); path != "" {
		if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			reg = tool.WrapWithAudit(reg, f)
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
	return ag, nil
}
