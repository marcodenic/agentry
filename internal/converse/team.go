package converse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"gopkg.in/yaml.v3"
)

func contextWithTeam(ctx context.Context, t *Team) context.Context {
	return team.WithContext(ctx, t)
}

// TeamFromContext extracts a Team pointer if present.
func TeamFromContext(ctx context.Context) (*Team, bool) {
	caller, ok := team.FromContext(ctx)
	if !ok {
		return nil, false
	}
	t, ok := caller.(*Team)
	return t, ok
}

// Team manages a multi-agent conversation step by step.
type Team struct {
	parent       *core.Agent
	agents       []*core.Agent
	names        []string
	agentsByName map[string]*core.Agent
	msg          string
	turn         int
	maxTurns     int
}

// RoleConfig represents a role configuration from YAML
type RoleConfig struct {
	Name   string   `yaml:"name"`
	Prompt string   `yaml:"prompt"`
	Tools  []string `yaml:"tools,omitempty"`
}

// loadRoleConfig loads a complete role configuration from the templates/roles directory
func loadRoleConfig(roleName string) (*RoleConfig, error) {
	// Try to find the templates/roles directory by searching up from the current directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	
	for dir := workDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		templatesDir := filepath.Join(dir, "templates", "roles")
		if _, err := os.Stat(templatesDir); err == nil {
			roleFile := filepath.Join(templatesDir, roleName+".yaml")
			if _, err := os.Stat(roleFile); err == nil {
				data, err := os.ReadFile(roleFile)
				if err != nil {
					return nil, err
				}
				
				var config RoleConfig
				if err := yaml.Unmarshal(data, &config); err != nil {
					return nil, err
				}
				
				return &config, nil
			}
		}
	}
	
	// Fallback to a generic config if role file not found
	return &RoleConfig{
		Name:   roleName,
		Prompt: fmt.Sprintf("You are a %s assistant. Help the user with tasks related to your specialization.", roleName),
		Tools:  []string{}, // Empty tools list for unknown roles
	}, nil
}

// loadRolePrompt loads a role-specific prompt from the templates/roles directory (deprecated, use loadRoleConfig)
func loadRolePrompt(roleName string) (string, error) {
	config, err := loadRoleConfig(roleName)
	if err != nil {
		return "", err
	}
	return config.Prompt, nil
}

// Add registers ag under name so it can be addressed via Call.
func (t *Team) Add(name string, ag *core.Agent) {
	t.agents = append(t.agents, ag)
	t.names = append(t.names, name)
	if t.agentsByName == nil {
		t.agentsByName = map[string]*core.Agent{}
	}
	t.agentsByName[name] = ag
}

// Agents returns the current set of agents in the team.
func (t *Team) Agents() []*core.Agent { return t.agents }

// Names returns the display names of the agents.
func (t *Team) Names() []string { return t.names }

// NewTeam spawns n sub-agents from parent ready to converse.
func NewTeam(parent *core.Agent, n int, topic string) (*Team, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be > 0")
	}
	if topic == "" {
		topic = "Hello agents, let's chat!"
	}

	shared := memory.NewInMemory()

	convRoute := parent.Route
	if rules, ok := parent.Route.(router.Rules); ok {
		cpy := make(router.Rules, len(rules))
		for i, r := range rules {
			cpy[i] = r
			cpy[i].Client = model.WithTemperature(r.Client, 0.7)
		}
		convRoute = cpy
	}

	agents := make([]*core.Agent, n)
	names := make([]string, n)
	byName := make(map[string]*core.Agent, n)
	for i := 0; i < n; i++ {
		ag := parent.Spawn()
		ag.Tracer = nil
		ag.Mem = shared
		ag.Route = convRoute
		agents[i] = ag
		name := fmt.Sprintf("Agent%d", i+1)
		names[i] = name
		byName[name] = ag
	}
	return &Team{parent: parent, agents: agents, names: names, agentsByName: byName, msg: topic, maxTurns: maxTurns}, nil
}

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
	ag, ok := t.agentsByName[name]
	if !ok {
		ag, _ = t.AddAgent(name)
	}
	ctx = contextWithTeam(ctx, t)
	return runAgent(ctx, ag, input, name, t.names)
}

// AddAgent spawns a new agent and joins it to the team. The returned agent and
// name can be used by callers. A default name is generated when none is
// provided.
func (t *Team) AddAgent(name string) (*core.Agent, string) {
	ag := t.parent.Spawn()
	ag.Tracer = nil
	ag.Mem = t.agents[0].Mem
	ag.Route = t.agents[0].Route
	
	// Set higher iteration limit for specialized agents
	ag.MaxIterations = 100 // Much higher limit for specialized agents
	
	// DEBUG: Show what we're doing
	fmt.Printf("🔄 DEBUG: AddAgent called with name: %s\n", name)
	fmt.Printf("🔄 DEBUG: Set MaxIterations to %d for %s\n", ag.MaxIterations, name)
	if len(ag.Prompt) > 100 {
		fmt.Printf("🔄 DEBUG: Original prompt (first 100 chars): %s...\n", ag.Prompt[:100])
	} else {
		fmt.Printf("🔄 DEBUG: Original prompt: %s\n", ag.Prompt)
	}
	
	// Load role-specific configuration for the named agent
	if roleConfig, err := loadRoleConfig(name); err == nil {
		if len(roleConfig.Prompt) > 100 {
			fmt.Printf("🔄 DEBUG: Loaded role prompt (first 100 chars): %s...\n", roleConfig.Prompt[:100])
		} else {
			fmt.Printf("🔄 DEBUG: Loaded role prompt: %s\n", roleConfig.Prompt)
		}
		ag.Prompt = roleConfig.Prompt
		
		// Create filtered tool registry based on role config
		if len(roleConfig.Tools) > 0 {
			filteredTools := make(tool.Registry)
			fmt.Printf("🔄 DEBUG: Filtering tools for %s. Allowed tools: %v\n", name, roleConfig.Tools)
			
			for _, toolName := range roleConfig.Tools {
				if t, ok := t.parent.Tools.Use(toolName); ok {
					filteredTools[toolName] = t
					fmt.Printf("🔄 DEBUG: Added tool '%s' to %s\n", toolName, name)
				} else {
					fmt.Printf("🔄 DEBUG: Tool '%s' not found in parent registry for %s\n", toolName, name)
				}
			}
			
			ag.Tools = filteredTools
			fmt.Printf("🔄 DEBUG: Agent %s now has %d tools (was %d)\n", name, len(ag.Tools), len(t.parent.Tools))
		} else {
			fmt.Printf("🔄 DEBUG: No tools specified for %s, keeping all parent tools\n", name)
		}
	} else {
		fmt.Printf("🔄 DEBUG: Failed to load role config: %v\n", err)
	}
	// If loading fails, keep the inherited prompt and tools as fallback
	
	if name == "" {
		name = fmt.Sprintf("Agent%d", len(t.agents)+1)
	}
	t.agents = append(t.agents, ag)
	t.names = append(t.names, name)
	t.agentsByName[name] = ag
	return ag, name
}
