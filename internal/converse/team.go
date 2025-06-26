package converse

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"gopkg.in/yaml.v3"
)

// agentNameRegex defines valid agent name pattern: starts with letter, contains letters, numbers, underscores, hyphens
var agentNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// isValidAgentName checks if an agent name follows the required conventions
func isValidAgentName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	return agentNameRegex.MatchString(name)
}

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
	Name        string   `yaml:"name"`
	Prompt      string   `yaml:"prompt"`
	Tools       []string `yaml:"tools,omitempty"`       // Legacy support
	Commands    []string `yaml:"commands,omitempty"`    // New semantic commands
	Builtins    []string `yaml:"builtins,omitempty"`    // Allowed builtin tools
	Personality string   `yaml:"personality,omitempty"` // For template substitution
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

// NewTeamContext creates a Team context ready for agents to be added dynamically.
// Unlike NewTeam, this does not pre-spawn any agents - they can be added later
// via AddAgent or other team spawning mechanisms.
func NewTeamContext(parent *core.Agent) (*Team, error) {
	return &Team{
		parent:       parent,
		agents:       make([]*core.Agent, 0), // Start with empty agent list
		names:        make([]string, 0),      // Start with empty names list
		agentsByName: make(map[string]*core.Agent), // Start with empty map
		msg:          "Hello agents, let's chat!", // Default initial message
		maxTurns:     maxTurns,
	}, nil
}

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
	// Check if the name is a tool - tools should not be created as agents
	if tool.IsBuiltinTool(name) {
		return "", fmt.Errorf("cannot create agent with tool name '%s': tool names are reserved", name)
	}
	
	ag, ok := t.agentsByName[name]
	if !ok {
		// Additional validation: enforce agent naming conventions
		if !isValidAgentName(name) {
			return "", fmt.Errorf("invalid agent name '%s': agent names must start with a letter and contain only letters, numbers, underscores, and hyphens", name)
		}
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
		
		// Inject platform-specific guidance with filtered commands
		if len(allowedCommands) > 0 || len(allowedBuiltins) > 0 {
			ag.Prompt = core.InjectPlatformContext(prompt, allowedCommands, allowedBuiltins)
		} else {
			ag.Prompt = prompt
		}
		
		// Create filtered tool registry based on builtins (if specified)
		if len(allowedBuiltins) > 0 {
			filteredTools := make(tool.Registry)
			
			for _, toolName := range allowedBuiltins {
				if t, ok := t.parent.Tools.Use(toolName); ok {
					filteredTools[toolName] = t
				}
			}
			
			ag.Tools = filteredTools
		}
		// If no builtins specified, inherit all tools from parent
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

// mapLegacyToolsToCommands converts old tool names to semantic commands for backward compatibility
func mapLegacyToolsToCommands(legacyTools []string) []string {
	toolMap := map[string][]string{
		"bash":       {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"powershell": {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"cmd":        {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"sh":         {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"ls":         {"list"},
		"view":       {"view"},
		"read":       {"view"},
		"write":      {"write"},
		"edit":       {"write"},
		"patch":      {"write"},
		"grep":       {"search"},
		"find":       {"find"},
		"fetch":      {}, // fetch is a builtin, not a semantic command
	}
	
	commandSet := make(map[string]bool)
	for _, tool := range legacyTools {
		if commands, exists := toolMap[tool]; exists {
			for _, cmd := range commands {
				commandSet[cmd] = true
			}
		}
	}
	
	var result []string
	for cmd := range commandSet {
		result = append(result, cmd)
	}
	
	return result
}
