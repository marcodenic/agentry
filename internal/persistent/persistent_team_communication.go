package persistent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SendMessage sends a message to a persistent agent via HTTP
func (pt *PersistentTeam) SendMessage(ctx context.Context, toAgentID, message string) (string, error) {
	agent, exists := pt.GetAgent(toAgentID)
	if !exists {
		return "", fmt.Errorf("agent %s not found", toAgentID)
	}

	// Send HTTP request to agent
	url := fmt.Sprintf("http://localhost:%d/message", agent.Port)
	
	messageData := map[string]string{
		"from":    "coordinator",
		"content": message,
		"type":    "task",
	}

	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Make HTTP request with JSON payload
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// For now, return success
	return "Message sent successfully", nil
}

// Call implements the team.Caller interface for compatibility with existing system
// This is the integration point where ephemeral delegation becomes persistent agents
func (pt *PersistentTeam) Call(ctx context.Context, agentID, input string) (string, error) {
	pt.mutex.RLock()
	agent, exists := pt.agents[agentID]
	pt.mutex.RUnlock()

	if !exists {
		// Agent doesn't exist yet - spawn it as a persistent agent
		// Try to determine role from agentID (coder, writer, tester, etc.)
		role := agentID
		if role == "" {
			role = "general"
		}

		var err error
		agent, err = pt.SpawnAgent(ctx, agentID, role)
		if err != nil {
			return "", fmt.Errorf("failed to spawn persistent agent %s: %w", agentID, err)
		}
		
		fmt.Printf("✅ Spawned persistent agent: %s (port %d)\n", agentID, agent.Port)
	}

	// Send task to persistent agent via HTTP
	return pt.SendMessage(ctx, agentID, input)
}
