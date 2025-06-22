package main

import (
	"errors"
	"fmt"

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
	if cfg.Store != "" {
		s, err := memstore.NewSQLite(cfg.Store)
		if err != nil {
			return nil, err
		}
		store = s
	}

	ag := core.New(rules, reg, memory.NewInMemory(), store, nil)
	return ag, nil
}
