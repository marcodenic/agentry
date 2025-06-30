package converse

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

// Step advances the conversation by one turn and returns the agent index and output.
func (t *Team) Step(ctx context.Context) (int, string, error) {
	if t.turn >= t.maxTurns {
		return -1, "", errors.New("max turns reached")
	}
	ctx = contextWithTeam(ctx, t)
	idx := t.turn % len(t.agents)
	out, err := runAgent(ctx, t.agents[idx], t.msg, t.names[idx], t.names)
	if err != nil {
		return idx, "", err
	}
	t.msg = out
	t.turn++
	return idx, out, nil
}

// ErrUnknownAgent is returned when Call is invoked with a name that doesn't exist.
var ErrUnknownAgent = errors.New("unknown agent")

// Call runs the named agent with the provided input once.
func (t *Team) Call(ctx context.Context, name, input string) (string, error) {
	debug.Printf("Team.Call invoked for agent '%s' with input: %s", name, input[:min(100, len(input))])

	// Check if the name is a tool - tools should not be created as agents
	if tool.IsBuiltinTool(name) {
		debug.Printf("Rejecting tool name '%s' as agent name", name)
		// Provide a helpful error message with suggestions for proper agent names
		suggestions := []string{"coder", "researcher", "analyst", "writer", "planner", "tester", "devops"}
		return "", fmt.Errorf("cannot create agent with tool name '%s': tool names are reserved. Use proper agent names like: %s",
			name, strings.Join(suggestions, ", "))
	}

	ag, ok := t.agentsByName[name]
	if !ok {
		debug.Printf("Agent '%s' not found, creating new agent", name)
		// Additional validation: enforce agent naming conventions
		if !isValidAgentName(name) {
			return "", fmt.Errorf("invalid agent name '%s': agent names must start with a letter and contain only letters, numbers, underscores, and hyphens", name)
		}
		ag, _ = t.AddAgent(name)
	} else {
		debug.Printf("Using existing agent '%s'", name)
	}
	ctx = contextWithTeam(ctx, t)
	return runAgent(ctx, ag, input, name, t.names)
}

// AddAgent spawns a new agent and joins it to the team. The returned agent and
// name can be used by callers. A default name is generated when none is
// provided.
func (t *Team) AddAgent(name string) (*core.Agent, string) {
	debug.Printf("Creating agent '%s'", name)

	ag := t.parent.Spawn()
	ag.Tracer = nil
	// Debug: Log parent agent information
	debug.Printf("Parent agent prompt length: %d chars", len(t.parent.Prompt))
	debug.Printf("Parent agent tools: %v", getToolNames(t.parent.Tools))
	debug.Printf("Spawned agent initial prompt: %s", ag.Prompt[:min(100, len(ag.Prompt))])

	// Set up shared memory and routing - use existing from first agent or create new ones
	if len(t.agents) > 0 {
		// Use existing shared memory and routing from the first agent
		ag.Mem = t.agents[0].Mem
		ag.Route = t.agents[0].Route
	} else {
		// This is the first agent being added, set up shared memory and routing
		shared := memory.NewInMemory()
		ag.Mem = shared

		convRoute := t.parent.Route
		if rules, ok := t.parent.Route.(router.Rules); ok {
			cpy := make(router.Rules, len(rules))
			for i, r := range rules {
				cpy[i] = r
				cpy[i].Client = model.WithTemperature(r.Client, 0.7)
			}
			convRoute = cpy
		}
		ag.Route = convRoute
	}

	// Set higher iteration limit for specialized agents
	ag.MaxIterations = 100 // Much higher limit for specialized agents

	// Load role-specific configuration for the named agent
	if roleConfig, err := loadRoleConfig(name); err == nil {
		debug.Printf("Loaded role config for '%s': prompt=%d chars, commands=%v, builtins=%v",
			name, len(roleConfig.Prompt), roleConfig.Commands, roleConfig.Builtins)

		// Apply template substitution if personality is provided
		prompt := roleConfig.Prompt
		if roleConfig.Personality != "" {
			prompt = strings.ReplaceAll(prompt, "{{personality}}", roleConfig.Personality)
		}

		// Determine which command and builtin lists to use
		var allowedCommands []string
		var allowedBuiltins []string

		// Use new semantic commands if available, otherwise fall back to legacy tools
		if len(roleConfig.Commands) > 0 {
			allowedCommands = roleConfig.Commands
		} else if len(roleConfig.Tools) > 0 {
			// Legacy support: map old tool names to semantic commands
			allowedCommands = mapLegacyToolsToCommands(roleConfig.Tools)
		}

		if len(roleConfig.Builtins) > 0 {
			allowedBuiltins = roleConfig.Builtins
		}

		debug.Printf("Agent '%s' allowed commands: %v, allowed builtins: %v",
			name, allowedCommands, allowedBuiltins)

		// Inject platform-specific guidance with filtered commands
		if len(allowedCommands) > 0 || len(allowedBuiltins) > 0 {
			ag.Prompt = core.InjectPlatformContext(prompt, allowedCommands, allowedBuiltins)
		} else {
			ag.Prompt = prompt
		}

		debug.Printf("Agent '%s' final prompt length: %d chars", name, len(ag.Prompt))
		debug.Printf("Agent '%s' final prompt preview: %s", name, ag.Prompt[:min(200, len(ag.Prompt))])

		// Create filtered tool registry based on builtins (if specified)
		if len(allowedBuiltins) > 0 {
			filteredTools := make(tool.Registry)

			for _, toolName := range allowedBuiltins {
				if t, ok := t.parent.Tools.Use(toolName); ok {
					filteredTools[toolName] = t
					debug.Printf("Agent '%s' granted builtin tool: %s", name, toolName)
				} else {
					debug.Printf("Agent '%s' requested unknown builtin tool: %s", name, toolName)
				}
			}

			ag.Tools = filteredTools
			debug.Printf("Agent '%s' final tools: %v", name, getToolNames(ag.Tools))
		} else {
			debug.Printf("Agent '%s' inheriting all parent tools: %v", name, getToolNames(ag.Tools))
		}
		// If no builtins specified, inherit all tools from parent
	} else {
		debug.Printf("Failed to load role config for '%s': %v - using default prompt", name, err)
	}
	// If loading fails, keep the inherited prompt and tools as fallback

	if name == "" {
		name = fmt.Sprintf("Agent%d", len(t.agents)+1)
	}
	t.agents = append(t.agents, ag)
	t.names = append(t.names, name)
	t.agentsByName[name] = ag

	// Debug logging
	debug.Printf("AddAgent: name=%s, agent=%v, prompt=%q, tools=%v", name, ag, ag.Prompt, ag.Tools)

	return ag, name
}
