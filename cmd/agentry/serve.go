package main

import (
	"context"
	"fmt"
	"github.com/marcodenic/agentry/internal/config"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/policy"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/session"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func runServe(args []string) {
	opts, _ := parseCommon("serve", args)
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}
	applyOverrides(cfg, opts)
	ag, err := buildAgent(cfg)
	if err != nil {
		panic(err)
	}
	if opts.maxIter > 0 {
		ag.MaxIterations = opts.maxIter
	}
	if dur, err := time.ParseDuration(cfg.SessionTTL); err == nil && dur > 0 {
		if cl, ok := ag.Store.(memstore.Cleaner); ok {
			interval := time.Hour
			if cfg.SessionGCInterval != "" {
				if iv, err := time.ParseDuration(cfg.SessionGCInterval); err == nil && iv > 0 {
					interval = iv
				}
			}
			session.Start(context.Background(), cl, dur, interval)
		}
	}
	if opts.ckptID != "" {
		ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(opts.ckptID))
		_ = ag.Resume(context.Background())
	}
	if opts.resumeID != "" {
		_ = ag.LoadState(context.Background(), opts.resumeID)
	}
	if cfg.Collector != "" {
		if _, err := trace.Init(cfg.Collector); err != nil {
			fmt.Printf("trace init: %v\n", err)
		}
		ag.Tracer = trace.NewOTel()
	}
	agents := map[string]*core.Agent{"default": ag}
	port := cfg.Port
	if opts.port != "" {
		port = opts.port
	}
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Serving HTTP on :%s\n", port)
	ap := policy.Manager{Prompt: policy.CLIPrompt}
	if err := server.Serve(port, agents, cfg.Metrics, opts.saveID, opts.resumeID, ap); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
