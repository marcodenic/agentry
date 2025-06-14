package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

// buildAgent constructs an Agent from configuration.
func buildAgent(cfg *config.File) (*core.Agent, error) {
	reg := tool.Registry{
		"echo": tool.New("echo", "Repeats the input", func(ctx context.Context, args map[string]any) (string, error) {
			txt, _ := args["text"].(string)
			return txt, nil
		}),
	}
	for _, m := range cfg.Tools {
		if m.Name == "echo" {
			continue
		}
		tl, err := tool.FromManifest(m)
		if err != nil {
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
		rules = append(rules, router.Rule{IfContains: rr.IfContains, Client: c})
	}
	if len(rules) == 0 {
		rules = router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
	}

	ag := core.New(rules, reg, memory.NewInMemory(), nil)
	return ag, nil
}
