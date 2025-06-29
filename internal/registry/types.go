package registry

import (
	"context"
	"time"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	StatusUnknown     AgentStatus = "unknown"
	StatusStarting    AgentStatus = "starting"
	StatusRunning     AgentStatus = "running"
	StatusIdle        AgentStatus = "idle"
	StatusBusy        AgentStatus = "busy"
	StatusWorking     AgentStatus = "working"
	StatusStopping    AgentStatus = "stopping"
	StatusStopped     AgentStatus = "stopped"
	StatusError       AgentStatus = "error"
	StatusShutdown    AgentStatus = "shutdown"
	StatusUnreachable AgentStatus = "unreachable"
)

// AgentInfo represents information about a registered agent
type AgentInfo struct {
	ID           string            `json:"id"`
	Port         int               `json:"port"`         // TCP port for localhost communication
	PID          int               `json:"pid"`          // Process ID for health checking
	Capabilities []string          `json:"capabilities"`
	Endpoint     string            `json:"endpoint"`     // "localhost:9001"
	Status       AgentStatus       `json:"status"`
	Metadata     map[string]string `json:"metadata"`
	LastSeen     time.Time         `json:"last_seen"`
	RegisteredAt time.Time         `json:"registered_at"`
	SessionID    string            `json:"session_id,omitempty"`
	Role         string            `json:"role,omitempty"`
	Version      string            `json:"version,omitempty"`
}

// HealthMetrics represents health and performance metrics for an agent
type HealthMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    int64   `json:"memory_usage"`
	TasksCompleted int64   `json:"tasks_completed"`
	TasksActive    int     `json:"tasks_active"`
	ErrorCount     int64   `json:"error_count"`
	Uptime         int64   `json:"uptime_seconds"`
}

// AgentRegistry defines the interface for agent registration and discovery
type AgentRegistry interface {
	// RegisterAgent registers a new agent with the registry
	RegisterAgent(ctx context.Context, info *AgentInfo) error
	
	// DeregisterAgent removes an agent from the registry
	DeregisterAgent(ctx context.Context, agentID string) error
	
	// UpdateAgent updates agent information
	UpdateAgent(ctx context.Context, agentID string, info *AgentInfo) error
	
	// GetAgent retrieves information about a specific agent
	GetAgent(ctx context.Context, agentID string) (*AgentInfo, error)
	
	// ListAllAgents returns all registered agents
	ListAllAgents(ctx context.Context) ([]*AgentInfo, error)
	
	// FindAgents finds agents with specific capabilities
	FindAgents(ctx context.Context, capabilities []string) ([]*AgentInfo, error)
	
	// UpdateAgentStatus updates the status of an agent
	UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error
	
	// UpdateAgentHealth updates health metrics for an agent
	UpdateAgentHealth(ctx context.Context, agentID string, health *HealthMetrics) error
	
	// GetAgentHealth retrieves health metrics for an agent
	GetAgentHealth(ctx context.Context, agentID string) (*HealthMetrics, error)
	
	// Heartbeat updates the last seen time for an agent
	Heartbeat(ctx context.Context, agentID string) error
	
	// Close closes the registry and any underlying resources
	Close() error
}

// RegistryEvent represents events that occur in the registry
type RegistryEvent struct {
	Type      string            `json:"type"`
	AgentID   string            `json:"agent_id"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]string `json:"data,omitempty"`
}

const (
	EventAgentRegistered   = "agent_registered"
	EventAgentDeregistered = "agent_deregistered"
	EventAgentStatusChange = "agent_status_change"
	EventAgentHeartbeat    = "agent_heartbeat"
	EventAgentUnreachable  = "agent_unreachable"
)

// EventSubscriber defines the interface for receiving registry events
type EventSubscriber interface {
	OnEvent(event *RegistryEvent) error
}

// PortRange defines a range of ports for agent communication
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
