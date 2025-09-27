package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

		// Provide available roles to the sectionized prompt as a dedicated <agents> section
		if ag.Prompt != "" {
			availableRoles := teamCtx.AvailableRoleNames()
			// Deterministic order
			sort.Strings(availableRoles)
			var sb strings.Builder
			sb.WriteString("AVAILABLE AGENTS: You can delegate tasks to these specialized agents using the 'agent' tool:\n\n")
			for _, role := range availableRoles {
				if role == "agent_0" { // don't list ourselves
					continue
				}
				sb.WriteString(role)
				sb.WriteString("\n")
			}
			sb.WriteString("\nExample delegation: {\"agent\": \"coder\", \"input\": \"create a hello world program\"}")
			if ag.Vars == nil {
				ag.Vars = map[string]string{}
			}
			ag.Vars["AGENTS_SECTION"] = sb.String()
			debug.Printf("Agent 0 agents section populated with %d roles", len(availableRoles))
		}

		// Register the agent delegation tool to replace the placeholder
		teamCtx.RegisterAgentTool(ag.Tools)
		debug.Printf("Agent delegation tool registered")
	}

	debug.Printf("After agent_0 config: agent has %d tools", len(ag.Tools))

	// No iteration cap
	col := trace.NewCollector(nil)
	ag.Tracer = col

	// Create context with team for coordination tools
	ctx := context.Background()
	if teamCtx != nil {
		ctx = team.WithContext(ctx, teamCtx)
		debug.Printf("Team context attached to execution context")
	}
	debug.Printf("Running Agent 0 with prompt length=%d", len(prompt))

	// Show actual useful information about what's happening
	taskPreview := prompt
	if len(prompt) > 100 {
		taskPreview = prompt[:100] + "..."
	}
	fmt.Fprintf(os.Stderr, "🤖 Agent 0 (%s) analyzing task: \"%s\"\n",
		ag.ModelName,
		taskPreview)

	if teamCtx != nil {
		availableAgents := teamCtx.AvailableRoleNames()
		fmt.Fprintf(os.Stderr, "� Available agents for delegation: %v\n", availableAgents)
	}

	out, err := ag.Run(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ ERR: %v\n", err)
		os.Exit(1)
	}

	sum := trace.Analyze(prompt, col.Events())

	// Show completion status
	fmt.Fprintf(os.Stderr, "✅ Task completed successfully!\n")

	// Print the model output to stdout
	fmt.Println(out)

	// Print usage summary to stderr (always show, not just in debug)
	if sum.TotalTokens > 0 {
		fmt.Fprintf(os.Stderr, "📊 Usage: %d input + %d output = %d total tokens, cost: $%.6f\n",
			sum.InputTokens, sum.OutputTokens, sum.TotalTokens, sum.Cost)
	}
}
