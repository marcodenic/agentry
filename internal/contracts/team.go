package contracts

import "context"

// TeamContextKey is used to store a team implementation in context for builtins.
// The concrete value should implement the TeamService interface.
var TeamContextKey = struct{ key string }{"agentry.team"}

// AgentNameContextKey provides the current agent's logical name (e.g., "agent_0" or role name)
// when running within a Team. Team sets this value before invoking the agent.
var AgentNameContextKey = struct{ key string }{"agentry.agent-name"}

// TeamService defines the contract for team coordination services.
// This interface breaks import cycles between tool and team packages.
type TeamService interface {
	// Agent Discovery
	SpawnedAgentNames() []string  // Currently running agent instances
	AvailableRoleNames() []string // Role names from configuration files

	// Task Delegation
	DelegateTask(ctx context.Context, role, task string) (string, error)

	// Communication
	SendMessage(ctx context.Context, from, to, message string) error
	GetInbox(agentID string) []map[string]interface{}
	MarkInboxRead(agentID string)

	// Coordination
	GetCoordinationSummary() string
	GetCoordinationHistory(limit int) []string

	// Shared Memory
	GetSharedData(key string) (interface{}, bool)
	SetSharedData(key string, value interface{})
	GetAllSharedData() map[string]interface{}

	// Help System
	RequestHelp(ctx context.Context, agentID, description, preferredHelper string) error
}
