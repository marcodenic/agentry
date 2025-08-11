package main

import (
	"fmt"
	"os"
	"strings"

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

	// If cmd starts with "-", treat it as flags for the default action
	// Look for the actual command or prompt after the flags
	if strings.HasPrefix(cmd, "-") {
		// Parse all arguments to find the actual command or prompt
		allArgs := os.Args[1:]
		actualCmd := ""
		actualArgs := []string{}
		
		// Find the first non-flag argument
		for i := 0; i < len(allArgs); i++ {
			arg := allArgs[i]
			if strings.HasPrefix(arg, "-") {
				// Skip flag and its value
				if i+1 < len(allArgs) && !strings.HasPrefix(allArgs[i+1], "-") {
					i++ // Skip the flag value
				}
			} else {
				// This is the actual command or prompt
				actualCmd = arg
				actualArgs = allArgs[:i] // All flags before this
				actualArgs = append(actualArgs, allArgs[i+1:]...) // Plus any remaining args
				break
			}
		}
		
		// If no command found after flags, default to TUI
		if actualCmd == "" {
			runTui(allArgs)
			return
		}
		
		// Check if it's a known command
		switch actualCmd {
		case "tui":
			runTui(actualArgs)
			return
		case "eval", "test":
			runEval(actualArgs)
			return
		// Add other commands as needed
		default:
			// It's a prompt, run it
			runPrompt(actualCmd, actualArgs)
			return
		}
	}

	switch cmd {
	case "chat":
		// Deprecated: chat mode is now an alias for the default TUI.
		fmt.Println("[deprecation] 'agentry chat' is deprecated. Use 'agentry' to launch the TUI.")
		runTui(args)
	case "dev":
		// Deprecated: dev mode is now an alias for the default TUI (use flags for debugging).
		fmt.Println("[deprecation] 'agentry dev' is deprecated. Use 'agentry' (TUI) with appropriate flags.")
		runTui(args)
	case "eval", "test":
		runEval(args)

	case "tui":
		runTui(args)
	case "invoke":
		runInvokeCmd(args)
	case "team":
		runTeamCmd(args)
	case "memory":
		runMemoryCmd(args)
	case "cost":
		runCostCmd(args)
	case "pprof":
		runPProfCmd(args)
	case "tool":
		runToolCmd(args)
	case "analyze":
		runAnalyzeCmd(args)
	case "refresh-models":
		runRefreshModelsCmd(args)
	case "version":
		fmt.Printf("agentry %s\n", agentry.Version)
	case "help", "-h", "--help":
		showHelp()
	default:
		runPrompt(cmd, args)
	}
}

func showHelp() {
	fmt.Printf(`agentry - AI Agent Coordination Platform

Usage:
	agentry <command> [options]

Commands:
	tui            Terminal UI mode (default when no command is provided)
	chat           [deprecated] Alias to TUI; use 'agentry' instead
	dev            [deprecated] Alias to TUI; use 'agentry' with flags instead
  eval, test     Run evaluations/tests
  cost           Analyze cost from trace logs
  pprof          Profiling utilities
  tool           Tool management
  analyze        Analyze trace files
	refresh-models Download and cache latest model pricing from models.dev
  version        Show version
  help           Show this help

Options:
  --config    Path to config file
  --theme     Theme override
  --version   Show version
  -v          Show version (short)
  --help      Show help

Examples:
	agentry                          # Start TUI (default)
	agentry tui                      # Start TUI explicitly
  agentry refresh-models           # Download latest model pricing
  agentry "create a hello world"   # Direct prompt
  agentry --version                # Show version
	agentry --help                   # Show help

Notes:
	- 'pprof' may require a build with tools enabled; see docs for diagnostics builds.
`)
}

func runToolCmd(_ []string) {
	fmt.Println("Tool command not implemented")
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
