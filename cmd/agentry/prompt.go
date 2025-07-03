package main

import (
	"context"
	"fmt"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/trace"
	"os"
	"strings"
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
	fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
}

func runPrompt(cmd string, args []string) {
	opts, _ := parseCommon(cmd, args)
	prompt := cmd
	if len(args) > 0 {
		prompt = cmd + " " + strings.Join(args, " ")
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
	if err := applyAgent0RoleConfig(ag); err != nil {
		fmt.Printf("Warning: Failed to apply agent_0 role configuration: %v\n", err)
	}
	fmt.Printf("🔧 After agent_0 config: agent has %d tools\n", len(ag.Tools))
	
	// FIX: Create team context for coordination capabilities (unified architecture)
	teamCtx, err := team.NewTeam(ag, 10, "")
	if err != nil {
		fmt.Printf("Warning: Failed to create team context: %v\n", err)
	} else {
		fmt.Printf("🔧 Team context created: Agent 0 now has coordination capabilities\n")
		// Register the agent delegation tool to replace the placeholder
		teamCtx.RegisterAgentTool(ag.Tools)
		fmt.Printf("🔧 Agent delegation tool registered with team\n")
	}
	
	if opts.maxIter > 0 {
		ag.MaxIterations = opts.maxIter
	}
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
	
	out, err := ag.Run(ctx, prompt)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
		os.Exit(1)
	}
	sum := trace.Analyze(prompt, col.Events())
	fmt.Println(out)
	fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
