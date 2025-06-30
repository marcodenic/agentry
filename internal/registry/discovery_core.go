package registry

import (
	"time"
)

// DiscoveryService provides intelligent agent discovery capabilities
type DiscoveryService struct {
	registry AgentRegistry
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(registry AgentRegistry) *DiscoveryService {
	return &DiscoveryService{
		registry: registry,
	}
}

// AgentScore represents an agent's suitability score for a task
type AgentScore struct {
	Agent *AgentInfo `json:"agent"`
	Score float64    `json:"score"`
}

// DiscoveryOptions contains options for agent discovery
type DiscoveryOptions struct {
	RequiredCapabilities []string  `json:"required_capabilities"`
	PreferredCapabilities []string  `json:"preferred_capabilities"`
	ExcludeAgents        []string  `json:"exclude_agents"`
	RequiredStatus       []AgentStatus `json:"required_status"`
	MaxResults           int       `json:"max_results"`
	SortBy               string    `json:"sort_by"` // "score", "load", "uptime"
}

// ClusterStatus represents the overall status of the agent cluster
type ClusterStatus struct {
	TotalAgents          int                    `json:"total_agents"`
	StatusCounts         map[AgentStatus]int    `json:"status_counts"`
	Capabilities         map[string]int         `json:"capabilities"`
	Roles               map[string]int         `json:"roles"`
	AverageUptime       time.Duration          `json:"average_uptime"`
	TotalTasksCompleted int64                  `json:"total_tasks_completed"`
	TotalErrors         int64                  `json:"total_errors"`
	LastUpdated         time.Time              `json:"last_updated"`
}
