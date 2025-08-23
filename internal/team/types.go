package team

import (
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
)

// Agent represents a team agent with metadata
type Agent struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Role      string            `json:"role,omitempty"`
	Agent     *core.Agent       `json:"-"`
	Port      int               `json:"port,omitempty"`
	Status    string            `json:"status"`
	StartedAt time.Time         `json:"started_at"`
	LastSeen  time.Time         `json:"last_seen"`
	Metadata  map[string]string `json:"metadata"`
	mutex     sync.RWMutex      `json:"-"`
}

// SetStatus updates the agent's status
func (a *Agent) SetStatus(status string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.Status = status
	a.LastSeen = time.Now()
}

// GetStatus returns the current agent status
func (a *Agent) GetStatus() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.Status
}

// Task represents a task assigned to an agent
type Task struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	AgentID   string            `json:"agent_id"`
	Input     string            `json:"input"`
	Result    string            `json:"result,omitempty"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Message represents a message between agents
type Message struct {
	ID        string            `json:"id"`
	From      string            `json:"from"`
	To        string            `json:"to"`
	Content   string            `json:"content"`
	Type      string            `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// RoleConfig represents configuration for an agent role
type RoleConfig struct {
	Name           string                `json:"name" yaml:"name"`
	Model          *config.ModelManifest `json:"model,omitempty" yaml:"model,omitempty"`
	Prompt         string                `json:"prompt" yaml:"prompt"`
	Tools          []string              `json:"tools,omitempty" yaml:"tools,omitempty"`
	RestrictedTools []string              `json:"restricted_tools,omitempty" yaml:"restricted_tools,omitempty"`
	Capabilities   []string              `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Metadata       map[string]string     `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PortRange defines a range of ports for agent communication
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// CoordinationEvent represents an event in agent coordination
type CoordinationEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "delegation", "message", "task_assign", "status_update"
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
