package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/audit"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/taskqueue"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func main() {
	cfgPath := flag.String("config", "examples/.agentry.yaml", "path to config file")
	concurrency := flag.Int("concurrency", 1, "max concurrent tasks")
	subject := flag.String("subject", "agentry.tasks", "NATS subject")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		panic(err)
	}
	ag, err := buildAgent(cfg)
	if err != nil {
		panic(err)
	}
	agents := map[string]*core.Agent{"default": ag}

	q, err := taskqueue.NewQueue(natsURL(), *subject)
	if err != nil {
		panic("NATS unavailable: " + err.Error())
	}
	defer q.Close()

	sem := make(chan struct{}, *concurrency)

	fmt.Println("Worker listening for tasks...")
	_, err = q.Subscribe(func(task taskqueue.Task) {
		if task.Type != "invoke" {
			return
		}
		payload, ok := task.Payload.(map[string]interface{})
		if !ok {
			fmt.Println("bad payload")
			return
		}
		agentID, _ := payload["agent_id"].(string)
		input, _ := payload["input"].(string)
		ag := agents[agentID]
		if ag == nil {
			fmt.Printf("unknown agent %s\n", agentID)
			return
		}
		sem <- struct{}{}
		go func(a *core.Agent, id, in string) {
			defer func() { <-sem }()
			if _, err := a.Run(context.Background(), in); err != nil {
				fmt.Printf("task %s error: %v\n", id, err)
			}
		}(ag, agentID, input)
	})
	if err != nil {
		panic(err)
	}
	select {}
}

func natsURL() string {
	if u := os.Getenv("NATS_URL"); u != "" {
		return u
	}
	return "nats://localhost:4222"
}

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
	if logWriter != nil {
		ag.Tracer = trace.NewJSONL(logWriter)
	}
	if cfg.MaxIterations > 0 {
		ag.MaxIterations = cfg.MaxIterations
	}
	return ag, nil
}
