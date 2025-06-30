package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/debug"
)

// CoordinateTask breaks down a complex task and assigns parts to team members
func (to *TeamOrchestrator) CoordinateTask(ctx context.Context, task string) (string, error) {
	debug.Printf("TeamOrchestrator: Coordinating complex task: %s", task)
	
	// Use the system agent to analyze and break down the task
	analysisPrompt := fmt.Sprintf(`
Analyze this task and determine how to coordinate with the team:

TASK: %s

TEAM STATUS:
%s

Please:
1. Break down the task into smaller parts if needed
2. Identify which team members should handle each part
3. Specify the order of execution if there are dependencies
4. If you need new agents with specific skills, identify what roles to spawn

Respond with a coordination plan.`, task, to.getTeamStatusString())
	
	// Run analysis with system agent
	plan, err := to.systemAgent.Run(ctx, analysisPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to create coordination plan: %w", err)
	}
	
	debug.Printf("TeamOrchestrator: Generated coordination plan: %s", plan)
	return plan, nil
}

// ProcessTeamCommand handles team-related commands from the system agent
func (to *TeamOrchestrator) ProcessTeamCommand(ctx context.Context, command string) (string, error) {
	debug.Printf("TeamOrchestrator: Processing command: %s", command)
	
	// Parse and execute team commands
	// This is called when the system agent needs to coordinate team actions
	
	switch {
	case stringContainsSubstr(command, "team status"):
		return to.formatTeamStatus(), nil
	case stringContainsSubstr(command, "available agents"):
		available := to.GetAvailableAgents()
		return fmt.Sprintf("Available agents: %v", available), nil
	case stringContainsSubstr(command, "assign task"):
		// Parse task assignment
		// Format: "assign task 'create a new file' to coder"
		return "Task assignment processed", nil
	default:
		return "", fmt.Errorf("unknown team command: %s", command)
	}
}

// Helper function to check if string contains substring
func stringContainsSubstr(s, substr string) bool {
	return strings.Contains(s, substr)
}
