package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/trace"
)

func runAnalyzeCmd(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: agentry analyze trace.log")
		return
	}
	sum, err := trace.AnalyzeFile(args[0])
	if err != nil {
		fmt.Println("analyze error:", err)
		os.Exit(1)
	}
	fmt.Printf("input tokens: %d, output tokens: %d, total tokens: %d, cost: $%.6f\n",
		sum.InputTokens, sum.OutputTokens, sum.TotalTokens, sum.Cost)
}

func runPrompt(cmd string, args []string) {
	// Use "agentry" as the flag set name for runPrompt, not the prompt text
	opts, remainingArgs := parseCommon("agentry", args)
	// The actual prompt is the cmd + any remaining args after flag parsing
	prompt := cmd
	if len(remainingArgs) > 0 {
		prompt = cmd + " " + strings.Join(remainingArgs, " ")
	}
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

	// Apply agent_0 role configuration to restrict tools
	fmt.Printf("🔧 Before agent_0 config: agent has %d tools\n", len(ag.Tools))

	// FIX: Create team context for coordination capabilities (unified architecture)
	// Load role configurations from include paths FIRST so Agent 0 can get proper config
	configDir := ""
	if opts.configPath != "" {
		configDir = filepath.Dir(opts.configPath)
	}
	teamCtx, err := team.NewTeamWithRoles(ag, 0, "", cfg.Include, configDir)
	if err != nil {
		fmt.Printf("Warning: Failed to create team context: %v\n", err)
	} else {
		fmt.Printf("🔧 Team context created: Agent 0 now has coordination capabilities\n")

		// Load Agent 0's proper role configuration directly
		agent0RolePath := "templates/roles/agent_0.yaml"
		if role, err := team.LoadRoleFromFile(agent0RolePath); err == nil {
			ag.Prompt = role.Prompt
			fmt.Printf("🔧 Agent 0 loaded proper role configuration from %s (prompt length: %d chars)\n", agent0RolePath, len(role.Prompt))
		} else {
			fmt.Printf("⚠️  Failed to load Agent 0 role from %s: %v\n", agent0RolePath, err)
		}

		// Register the agent delegation tool to replace the placeholder
		teamCtx.RegisterAgentTool(ag.Tools)
		fmt.Printf("🔧 Agent delegation tool registered with team\n")
	}

	fmt.Printf("🔧 After agent_0 config: agent has %d tools\n", len(ag.Tools))

	// No iteration cap
	if opts.resumeID != "" {
		_ = ag.LoadState(context.Background(), opts.resumeID)
	}
	col := trace.NewCollector(nil)
	ag.Tracer = col

	// Create context with team for coordination tools (matching chat mode)
	ctx := context.Background()
	if teamCtx != nil {
		ctx = team.WithContext(ctx, teamCtx)
		fmt.Printf("🔧 Team context attached to execution context\n")
	}

	fmt.Printf("🔧 Running Agent 0 with prompt: %q\n", prompt)
	out, err := ag.Run(ctx, prompt)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
		os.Exit(1)
	}
	sum := trace.Analyze(prompt, col.Events())
	fmt.Println(out)
	fmt.Printf("input tokens: %d, output tokens: %d, total tokens: %d, cost: $%.6f\n",
		sum.InputTokens, sum.OutputTokens, sum.TotalTokens, sum.Cost)
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
