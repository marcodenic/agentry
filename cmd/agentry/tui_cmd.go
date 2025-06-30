package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/tui"
)

func runTui(args []string) {
	// Enable TUI mode to suppress debug output
	debug.SetTUIMode(true)
	
	// Redirect stdout/stderr to prevent debug output from interfering with TUI
	origStdout := os.Stdout
	origStderr := os.Stderr
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	if err == nil {
		os.Stdout = devNull
		os.Stderr = devNull
		defer func() {
			devNull.Close()
			os.Stdout = origStdout
			os.Stderr = origStderr
		}()
	}
	
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

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel() // Cancel context to signal shutdown
	}()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithContext(ctx))
	if err := p.Start(); err != nil {
		panic(err)
	}

	cancel() // Ensure cleanup even if program exits normally
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
