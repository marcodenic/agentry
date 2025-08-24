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
	
	// If no arguments, start TUI
	if len(os.Args) < 2 {
		runTui([]string{})
		return
	}

	// Handle version flags first
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("agentry %s\n", agentry.Version)
		return
	}

	// Handle help
	if os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
		return
	}

	// Determine the command (first non-flag argument or infer from context)
	args := os.Args[1:]
	var command string
	var commandArgs []string
	
	// Find first non-flag argument to use as command
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			// Check if this looks like a known command
			switch arg {
			case "tui", "chat", "ask", "prompt", "refresh-models", "version":
				command = arg
				commandArgs = args[i+1:]
			default:
				// This is a direct prompt - everything from here is part of the prompt
				command = "prompt-direct"
				commandArgs = args[i:] // Include this arg and everything after
			}
			break
		}
	}
	
	// If we only found flags, default to TUI
	if command == "" {
		command = "tui"
		commandArgs = args
	}

	// Handle explicit commands
	switch command {
	case "tui":
		runTui(commandArgs)
	case "chat", "ask", "prompt":
		if len(commandArgs) == 0 {
			fmt.Println("Error: chat command requires a prompt")
			fmt.Println("Usage: agentry chat \"your prompt here\"")
			os.Exit(1)
		}
		runPrompt(strings.Join(commandArgs, " "), args[:len(args)-len(commandArgs)])
	case "refresh-models":
		runRefreshModelsCmd(commandArgs)
	case "version":
		fmt.Printf("agentry %s\n", agentry.Version)
	case "prompt-direct":
		// Direct prompt with all arguments and flags
		runPrompt(strings.Join(commandArgs, " "), args[:len(args)-len(commandArgs)])
	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		fmt.Println("Use 'agentry help' for usage information")
		os.Exit(1)
	}
}

func showHelp() {
	helpText := `Agentry - Multi-agent orchestrator for development tasks

USAGE:
  agentry [command] [flags] [arguments]

COMMANDS:
  (no command)           Start TUI interface (default)
  chat [prompt]          Interactive chat mode, optionally with initial prompt
  tui                    Start TUI interface (same as no command)
  refresh-models         Update model pricing data
  help                   Show this help message
  
  Direct prompt execution:
  agentry "quoted prompt"    Execute prompt directly with Agent 0
  agentry unquoted prompt    No quotes needed for simple prompts

FLAGS:
  --config PATH          Path to .agentry.yaml config file
  --theme THEME          Theme override (dark|light|auto)  
  --debug                Enable debug output
  --keybinds PATH        Path to custom keybindings JSON file
  --creds PATH           Path to credentials JSON file
  --mcp SERVERS          Comma-separated MCP server list
  --save-id ID           Save conversation state to this ID
  --resume-id ID         Load conversation state from this ID  
  --checkpoint-id ID     Checkpoint session ID
  --port PORT            HTTP server port
  --disable-tools        Disable tool filtering entirely
  --allow-tools TOOLS    Comma-separated list of additional tools to include
  --deny-tools TOOLS     Comma-separated list of tools to exclude
  --disable-context      Disable context pipeline
  --audit-log PATH       Path to audit log file

EXAMPLES:
  agentry                                    # Start TUI
  agentry fix the auth tests                 # Direct prompt (no quotes needed)
  agentry "complex prompt with & symbols"    # Quotes for special characters
  agentry chat hello there                   # Chat with initial prompt
  agentry --debug --theme dark analyze code # Debug mode with dark theme
  agentry tui --resume-id my-session         # Resume TUI session
  agentry refresh-models                     # Update model data

For more information, see PRODUCT.md or visit the project repository.
`
	fmt.Print(helpText)
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
