package core

import (
	"fmt"
)

// GetTeamStatus returns the current status of all team agents
func (to *TeamOrchestrator) GetTeamStatus() map[string]AgentStatus {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	status := make(map[string]AgentStatus)
	for name, agent := range to.agentStatus {
		status[name] = agent
	}
	
	return status
}

// GetSystemPrompt returns an enhanced system prompt that includes team context
func (to *TeamOrchestrator) GetSystemPrompt() string {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	teamStatus := "TEAM STATUS:\n"
	for name, status := range to.agentStatus {
		teamStatus += fmt.Sprintf("- %s (%s): %s", name, status.Role, status.State)
		if status.CurrentTask != "" {
			teamStatus += fmt.Sprintf(" - %s (%.0f%%)", status.CurrentTask, status.Progress*100)
		}
		teamStatus += "\n"
	}
	
	availableAgents := to.GetAvailableAgents()
	teamStatus += fmt.Sprintf("\nAVAILABLE AGENTS: %v\n", availableAgents)
	
	basePrompt := to.systemAgent.Prompt
	return fmt.Sprintf("%s\n\n%s\nYou can coordinate with team members using your tools:\n- Use the 'agent' tool to delegate tasks: {\"agent\": \"coder\", \"input\": \"create hello.py\"}\n- Use 'team_status' to check team state\n- Use 'assign_task' to formally assign tasks\n- Use 'send_message' to communicate with agents", basePrompt, teamStatus)
}

// getTeamStatusString returns a formatted string of team status
func (to *TeamOrchestrator) getTeamStatusString() string {
	status := ""
	for name, agent := range to.agentStatus {
		status += fmt.Sprintf("- %s (%s): %s", name, agent.Role, agent.State)
		if agent.CurrentTask != "" {
			status += fmt.Sprintf(" - %s", agent.CurrentTask)
		}
		status += "\n"
	}
	return status
}

// formatTeamStatus formats team status for display
func (to *TeamOrchestrator) formatTeamStatus() string {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	result := "Team Status Report:\n\n"
	for name, status := range to.agentStatus {
		result += fmt.Sprintf("Agent: %s (%s)\n", name, status.Role)
		result += fmt.Sprintf("  State: %s\n", status.State)
		if status.CurrentTask != "" {
			result += fmt.Sprintf("  Task: %s (%.0f%% complete)\n", status.CurrentTask, status.Progress*100)
		}
		result += fmt.Sprintf("  Last Update: %s\n", status.LastUpdate.Format("15:04:05"))
		if status.TokenCount > 0 {
			result += fmt.Sprintf("  Tokens Used: %d\n", status.TokenCount)
		}
		result += "\n"
	}
	
	return result
}
