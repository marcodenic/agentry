package core

import (
	"fmt"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

// AssignTask assigns a task to a specific agent
func (to *TeamOrchestrator) AssignTask(agentName, task string) error {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	// Check if agent exists and is available
	status, exists := to.agentStatus[agentName]
	if !exists {
		return fmt.Errorf("agent '%s' not found", agentName)
	}
	
	if status.State == "working" {
		return fmt.Errorf("agent '%s' is currently busy", agentName)
	}
	
	// Update agent status
	status.State = "working"
	status.CurrentTask = task
	status.Progress = 0.0
	status.LastUpdate = time.Now()
	to.agentStatus[agentName] = status
	
	// Store task
	to.agentTasks[agentName] = task
	
	// Send task message
	to.SendMessage("system", agentName, "task", task, map[string]interface{}{
		"assigned_at": time.Now(),
		"priority":    "normal",
	})
	
	debug.Printf("TeamOrchestrator: Assigned task to %s: %s", agentName, task)
	return nil
}

// CompleteTask marks a task as completed and stores the result
func (to *TeamOrchestrator) CompleteTask(agentName, result string) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	// Update agent status
	if status, exists := to.agentStatus[agentName]; exists {
		status.State = "idle"
		status.CurrentTask = ""
		status.Progress = 1.0
		status.LastUpdate = time.Now()
		to.agentStatus[agentName] = status
	}
	
	// Store result
	to.agentResults[agentName] = result
	
	// Notify system agent
	to.notifySystemAgent(fmt.Sprintf("Agent '%s' completed task: %s", agentName, result))
	
	debug.Printf("TeamOrchestrator: Task completed by %s", agentName)
}
