package main

import (
	"fmt"
	"os"

	agentry "github.com/marcodenic/agentry/internal"
	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/env"
)

func main() {
	env.Load()
	if len(os.Args) < 2 {
		runTui([]string{})
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	// Handle version flags first
	if cmd == "--version" || cmd == "-v" {
		fmt.Printf("agentry %s\n", agentry.Version)
		return
	}

	switch cmd {
	case "refresh-models":
		runRefreshModelsCmd(args)
	case "version":
		fmt.Printf("agentry %s\n", agentry.Version)
	case "help", "-h", "--help":
		showHelp()
	default:
		// Everything else is either TUI with flags or a direct prompt
		runPrompt(cmd, args)
	}
}

func showHelp() {
	fmt.Printf(`agentry - AI Agent Coordination Platform

Usage:
	agentry [flags]           Start TUI (Terminal UI)
	agentry "prompt text"     Execute prompt directly

Commands:
	refresh-models    Download and cache latest model pricing from models.dev
	version           Show version
	help              Show this help

Direct Prompt:
	agentry "create a hello world"      # Execute prompt directly
	agentry "fix the bug in main.go"    # Direct task execution

TUI Options:
	--config PATH     Path to config file (.agentry.yaml)
	--theme NAME      Theme override (dark, light, etc.)
	--save-id ID      Save conversation state to this ID
	--resume-id ID    Load conversation state from this ID
	--debug           Enable debug output

Debug Mode:
	AGENTRY_DEBUG=1 ./agentry "prompt"  # Enable debug output
	AGENTRY_DEBUG=1 ./agentry           # Debug TUI (logs to file)

Examples:
	agentry                              # Start TUI (default)
	agentry "write a README file"        # Direct prompt execution
	agentry --debug "test the code"      # Direct prompt with debug output
	agentry refresh-models               # Update model pricing

Notes:
	- Agent 0 automatically delegates to specialized agents through the 'agent' tool
	- All multi-agent coordination happens through delegation, not separate commands
	- Use --debug for detailed execution traces and diagnostics
`)
}

// Stub implementation for optional command if not present in this build.
func runRefreshModelsCmd(_ []string) {
	fmt.Println("Fetching latest model pricing/specs from models.dev ...")
	pt := cost.NewPricingTable()
	if err := pt.RefreshFromAPI(); err != nil {
		fmt.Printf("Failed to refresh: %v\n", err)
		os.Exit(1)
	}
	// Give a small summary
	models := pt.ListModels()
	fmt.Printf("Refreshed %d models and cached to your user cache dir (agentry/models_pricing.json)\n", len(models))
}
