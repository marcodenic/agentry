package core

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// TeamOrchestrator enhances Agent 0 with team coordination capabilities
type TeamOrchestrator struct {
	systemAgent   *Agent
	teamAgents    map[string]*Agent
	agentStatus   map[string]AgentStatus
	agentTasks    map[string]string
	agentResults  map[string]string
	messageQueue  []TeamMessage
	mutex         sync.RWMutex
}

// AgentStatus represents the current state of an agent
type AgentStatus struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	State       string    `json:"state"`       // "idle", "working", "waiting", "error"
	CurrentTask string    `json:"current_task"`
	Progress    float64   `json:"progress"`    // 0.0 to 1.0
	LastUpdate  time.Time `json:"last_update"`
	TokenCount  int       `json:"token_count"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
}

// TeamMessage represents communication between agents
type TeamMessage struct {
	ID        uuid.UUID `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`        // "" for broadcast
	Type      string    `json:"type"`      // "task", "status", "result", "question"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewTeamOrchestrator creates a new orchestrator with the system agent
func NewTeamOrchestrator(systemAgent *Agent) *TeamOrchestrator {
	return &TeamOrchestrator{
		systemAgent:  systemAgent,
		teamAgents:   make(map[string]*Agent),
		agentStatus:  make(map[string]AgentStatus),
		agentTasks:   make(map[string]string),
		agentResults: make(map[string]string),
		messageQueue: make([]TeamMessage, 0),
	}
}
