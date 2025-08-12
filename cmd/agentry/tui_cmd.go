package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/tui"
)

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func runTui(args []string) {
	// Load .env.local file if it exists
	if _, err := os.Stat(".env.local"); err == nil {
		if err := loadEnvFile(".env.local"); err != nil {
			fmt.Printf("Warning: failed to load .env.local: %v\n", err)
		}
	}

	// Enable TUI mode to suppress debug output
	debug.SetTUIMode(true)

	// Set environment variable to prevent logging interference
	os.Setenv("AGENTRY_TUI_MODE", "1")

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

	// No iteration cap
	if opts.ckptID != "" {
		ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(opts.ckptID))
		_ = ag.Resume(context.Background())
	}
	if opts.resumeID != "" {
		_ = ag.LoadState(context.Background(), opts.resumeID)
	}

	// Pass config information to TUI for role loading
	configDir := ""
	if opts.configPath != "" {
		configDir = filepath.Dir(opts.configPath)
	}
	model := tui.NewWithConfig(ag, cfg.Include, configDir)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel() // Cancel context to signal shutdown
	}()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	cancel() // Ensure cleanup even if program exits normally
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
