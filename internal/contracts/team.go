package contracts

import "context"

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
