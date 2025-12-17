package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/trace"
)

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
	// Apply iteration cap from flags (0 = unlimited)
	ag.MaxIter = opts.maxIter

	debug.Printf("Agent initialized with %d tools", len(ag.Tools))

	// Create team context - simplified architecture (no specialized agents)
	configDir := ""
	if opts.configPath != "" {
		configDir = filepath.Dir(opts.configPath)
	}
	teamCtx, err := team.NewTeamWithRoles(ag, 0, "", cfg.Include, configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create team context: %v\n", err)
	} else {
		// Register the sub-agent tool for parallel search operations
		teamCtx.RegisterAgentTool(ag.Tools)
		debug.Printf("Sub-agent tool registered for parallel search")
	}

	debug.Printf("Agent has %d tools available", len(ag.Tools))

	col := trace.NewCollector(nil)
	ag.Tracer = col

	// Create context with team for sub-agent tool access
	ctx := context.Background()
	if teamCtx != nil {
		ctx = team.WithContext(ctx, teamCtx)
	}

	// Show task preview
	taskPreview := prompt
	if len(prompt) > 100 {
		taskPreview = prompt[:100] + "..."
	}
	fmt.Fprintf(os.Stderr, "🤖 Agent (%s) starting: \"%s\"\n", ag.ModelName, taskPreview)

	out, err := ag.Run(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ ERR: %v\n", err)
		os.Exit(1)
	}

	sum := trace.Analyze(prompt, col.Events())

	fmt.Fprintf(os.Stderr, "✅ Task completed!\n")
	fmt.Println(out)

	if sum.TotalTokens > 0 {
		fmt.Fprintf(os.Stderr, "📊 Usage: %d input + %d output = %d tokens, $%.6f\n",
			sum.InputTokens, sum.OutputTokens, sum.TotalTokens, sum.Cost)
	}
}
