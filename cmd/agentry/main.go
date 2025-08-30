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
    
    // If no arguments, start TUI (default)
    if len(os.Args) < 2 {
        runTui([]string{})
        return
    }

	// Handle version and help flags specially (before parsing)
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("agentry %s\n", agentry.Version)
		return
	}
	if os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
		return
	}

	// Parse all arguments to separate global flags from command and its args
	args := os.Args[1:]
	opts, remainingArgs := parseCommon("agentry", args)
	
    // If no remaining args after flag parsing, start TUI (default)
    if len(remainingArgs) == 0 {
        runTui(args) // Pass original args to TUI for its own parsing
        return
    }

    // Determine the command (first remaining argument or infer from context)
    var command string
    var commandArgs []string

    // Recognized commands: tui, refresh-models, version. Deprecated aliases: chat/ask/prompt → direct prompt.
    switch remainingArgs[0] {
    case "tui", "refresh-models", "version":
        command = remainingArgs[0]
        commandArgs = remainingArgs[1:]
    case "chat", "ask", "prompt":
        // Deprecated aliases: treat everything after as a direct prompt
        command = "prompt-direct"
        commandArgs = remainingArgs[1:]
    default:
        // Direct prompt with all remaining args
        command = "prompt-direct"
        commandArgs = remainingArgs
    }

	// Handle explicit commands
	switch command {
    case "tui":
        runTui(commandArgs)
    case "refresh-models":
        runRefreshModelsCmd(commandArgs)
    case "version":
        fmt.Printf("agentry %s\n", agentry.Version)
    case "prompt-direct":
        // Direct prompt with all arguments
        runPromptWithOpts(strings.Join(commandArgs, " "), opts)
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
    (no command)         Start TUI interface (default)
  refresh-models       Update model pricing data
  help                 Show this help message
  
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
  --disable-tools        Disable tool filtering entirely (allow all tools)
  --allow-tools TOOLS    Restrict to only specified tools (comma-separated)
  --deny-tools TOOLS     Remove specific tools from available set (comma-separated)
  --disable-context      Disable context pipeline
  --audit-log PATH       Path to audit log file

EXAMPLES:
  agentry                                  # Start TUI (default)
  agentry fix the auth tests               # Direct prompt (no quotes needed)
  agentry "complex prompt with & symbols"  # Quotes for special characters
  agentry --debug analyze code             # Debug mode with direct prompt
  agentry --resume-id my-session           # Resume TUI session
  agentry refresh-models                   # Update model data
  
  Tool filtering examples:
  agentry --allow-tools echo,ping "test"           # Only echo and ping tools
  agentry --deny-tools bash,sh "safe operation"    # No shell access
  agentry --disable-tools "unrestricted access"    # All tools available

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
