package core

import (
	"fmt"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

// RegisterAgent adds an agent to the team
func (to *TeamOrchestrator) RegisterAgent(name string, agent *Agent, role string) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	to.teamAgents[name] = agent
	to.agentStatus[name] = AgentStatus{
		ID:         agent.ID,
		Name:       name,
		Role:       role,
		State:      "idle",
		LastUpdate: time.Now(),
		Progress:   0.0,
	}
	
	debug.Printf("TeamOrchestrator: Registered agent '%s' with role '%s'", name, role)
	
	// Notify system agent about new team member
	to.notifySystemAgent(fmt.Sprintf("Agent '%s' (%s) has joined the team", name, role))
}

// UpdateAgentStatus updates the status of a specific agent
func (to *TeamOrchestrator) UpdateAgentStatus(name, state, task string, progress float64) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	if status, exists := to.agentStatus[name]; exists {
		status.State = state
		status.CurrentTask = task
		status.Progress = progress
		status.LastUpdate = time.Now()
		to.agentStatus[name] = status
		
		debug.Printf("TeamOrchestrator: Updated %s status: %s (%.1f%%)", name, state, progress*100)
	}
}

// GetAvailableAgents returns a list of agents that are currently idle
func (to *TeamOrchestrator) GetAvailableAgents() []string {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	var available []string
	for name, status := range to.agentStatus {
		if status.State == "idle" {
			available = append(available, name)
		}
	}
	
	return available
}

// notifySystemAgent sends a notification to the system agent's memory
func (to *TeamOrchestrator) notifySystemAgent(message string) {
	// Add the notification to system agent's memory context
	// This allows Agent 0 to be aware of team activities
	debug.Printf("SystemAgent notification: %s", message)
}
