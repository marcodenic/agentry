package main

import (
	"fmt"
	"os"

	agentry "github.com/marcodenic/agentry/internal"
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
	case "chat":
		runChatMode(args)
	case "dev":
		runDev(args)
	case "eval", "test":
		runEval(args)

	case "tui":
		runTui(args)
	case "cost":
		runCostCmd(args)
	case "pprof":
		runPProfCmd(args)
	case "tool":
		runToolCmd(args)
	case "analyze":
		runAnalyzeCmd(args)
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
  chat        Interactive chat mode for natural language delegation
  tui         Terminal UI mode with rich interface (default)
  dev         Development REPL with tracing
  eval, test  Run evaluations/tests
  cost        Analyze cost from trace logs
  pprof       Profiling utilities
  tool        Tool management
  analyze     Analyze trace files
  version     Show version
  help        Show this help

Options:
  --config    Path to config file
  --theme     Theme override
  --version   Show version
  -v          Show version (short)
  --help      Show help

Examples:
  agentry                          # Start TUI (default)
  agentry chat                     # Start interactive chat
  agentry tui                      # Start TUI explicitly
  agentry dev                      # Start development REPL
  agentry "create a hello world"   # Direct prompt
  agentry --version                # Show version
  agentry --help                   # Show help
`)
}

func runToolCmd(args []string) {
	fmt.Println("Tool command not implemented")
}
