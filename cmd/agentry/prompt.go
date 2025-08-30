package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/marcodenic/agentry/internal/config"
    "github.com/marcodenic/agentry/internal/core"
    "github.com/marcodenic/agentry/internal/debug"
    "github.com/marcodenic/agentry/internal/team"
    "github.com/marcodenic/agentry/internal/trace"
)

func runPrompt(prompt string, args []string) {
	// Parse any flags that might be passed with the prompt
	opts, remainingArgs := parseCommon("agentry", args)
	
	// If there are remaining args, append them to the prompt
	if len(remainingArgs) > 0 {
		prompt = prompt + " " + strings.Join(remainingArgs, " ")
	}
	
	runPromptWithOpts(prompt, opts)
}

func runPromptWithOpts(prompt string, opts *commonOpts) {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
    fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	applyOverrides(cfg, opts)
	ag, err := buildAgent(cfg)
	if err != nil {
		panic(err)
	}

    // Debug: tool count before/after role configuration
    debug.Printf("Before agent_0 config: agent has %d tools", len(ag.Tools))

	// FIX: Create team context for coordination capabilities (unified architecture)
	// Load role configurations from include paths FIRST so Agent 0 can get proper config
	configDir := ""
	if opts.configPath != "" {
		configDir = filepath.Dir(opts.configPath)
	}
    teamCtx, err := team.NewTeamWithRoles(ag, 0, "", cfg.Include, configDir)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: Failed to create team context: %v\n", err)
    } else {
        debug.Printf("Team context created: Agent 0 has coordination capabilities")

		// Load Agent 0's proper role configuration directly
		agent0RolePath := "templates/roles/agent_0.yaml"
        if role, err := team.LoadRoleFromFile(agent0RolePath); err == nil {
            ag.Prompt = role.Prompt
            debug.Printf("Agent 0 loaded role configuration from %s (prompt length: %d chars)", agent0RolePath, len(role.Prompt))
        } else {
            debug.Printf("Failed to load Agent 0 role from %s: %v", agent0RolePath, err)
        }

		// CRITICAL: Enhance Agent0 prompt with available roles information
        if ag.Prompt != "" {
            availableRoles := teamCtx.AvailableRoleNames()
            ag.Prompt = core.InjectAvailableRoles(ag.Prompt, availableRoles)
            debug.Printf("Agent 0 enhanced with %d available roles", len(availableRoles))
        }

		// Register the agent delegation tool to replace the placeholder
        teamCtx.RegisterAgentTool(ag.Tools)
        debug.Printf("Agent delegation tool registered")
    }

    debug.Printf("After agent_0 config: agent has %d tools", len(ag.Tools))

	// No iteration cap
    if opts.resumeID != "" {
        _ = ag.LoadState(context.Background(), opts.resumeID)
    }
	col := trace.NewCollector(nil)
	ag.Tracer = col

    // Create context with team for coordination tools
	ctx := context.Background()
    if teamCtx != nil {
        ctx = team.WithContext(ctx, teamCtx)
        debug.Printf("Team context attached to execution context")
    }
    debug.Printf("Running Agent 0 with prompt length=%d", len(prompt))
	out, err := ag.Run(ctx, prompt)
    if err != nil {
        fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
        os.Exit(1)
    }
    sum := trace.Analyze(prompt, col.Events())
    // Print only the model output to stdout
    fmt.Println(out)
    // Print usage summary to stderr in debug mode
    if debug.DebugEnabled {
        fmt.Fprintf(os.Stderr, "input tokens: %d, output tokens: %d, total tokens: %d, cost: $%.6f\n",
            sum.InputTokens, sum.OutputTokens, sum.TotalTokens, sum.Cost)
    }
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
