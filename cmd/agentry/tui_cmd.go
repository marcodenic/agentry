package main

import (
	"context"
	"fmt"
	"github.com/marcodenic/agentry/internal/config"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/tui"
)

func runTui(args []string) {
	opts, _ := parseCommon("tui", args)
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
	if opts.ckptID != "" {
		ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(opts.ckptID))
		_ = ag.Resume(context.Background())
	}
	if opts.resumeID != "" {
		_ = ag.LoadState(context.Background(), opts.resumeID)
	}
	model := tui.New(ag)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		panic(err)
	}
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
