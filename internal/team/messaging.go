package team

import (
	"context"
	"fmt"
	"os"
	"time"
)

// SendMessageToAgent enables direct communication between agents
func (t *Team) SendMessageToAgent(ctx context.Context, fromAgentID, toAgentID, message string) error {
	t.mutex.RLock()
	_, fromExists := t.agentsByName[fromAgentID]
	_, toExists := t.agentsByName[toAgentID]
	t.mutex.RUnlock()

	if !fromExists {
		return fmt.Errorf("sender agent %s not found", fromAgentID)
	}
	if !toExists {
		return fmt.Errorf("recipient agent %s not found", toAgentID)
	}

	// Log the direct communication
	if !isTUI() {
		fmt.Fprintf(os.Stderr, "💬 DIRECT MESSAGE: %s → %s\n", fromAgentID, toAgentID)
		fmt.Fprintf(os.Stderr, "📝 Message: %s\n", message)
	}

	// Store in coordination events
	t.LogCoordinationEvent("direct_message", fromAgentID, toAgentID, message, map[string]interface{}{
		"message_type": "agent_to_agent",
		"timestamp":    time.Now(),
	})

	// Append to typed message history and mark unread
	t.mutex.Lock()
	t.messages = append(t.messages, Message{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		From:      fromAgentID,
		To:        toAgentID,
		Content:   message,
		Type:      "direct",
		Timestamp: time.Now(),
		Read:      false,
	})
	t.mutex.Unlock()

	return nil
}

// BroadcastToAllAgents sends a message to all agents
func (t *Team) BroadcastToAllAgents(ctx context.Context, fromAgentID, message string) error {
	agentNames := t.GetAgents()
	if !isTUI() {
		fmt.Fprintf(os.Stderr, "📢 BROADCAST from %s: %s\n", fromAgentID, message)
	}
	for _, agentName := range agentNames {
		if agentName != fromAgentID { // Don't send to self
			if err := t.SendMessageToAgent(ctx, fromAgentID, agentName, message); err != nil {
				if !isTUI() {
					fmt.Fprintf(os.Stderr, "❌ Failed to broadcast to %s: %v\n", agentName, err)
				}
			}
		}
	}
	return nil
}

// GetAgentInbox returns unread messages for an agent (as generic maps for tool consumption)
func (t *Team) GetAgentInbox(agentID string) []map[string]interface{} {
	// Build a view over typed messages for this agent
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	out := make([]map[string]interface{}, 0)
	for _, m := range t.messages {
		if m.To != agentID {
			continue
		}
		out = append(out, map[string]interface{}{
			"from":      m.From,
			"message":   m.Content,
			"timestamp": m.Timestamp,
			"read":      m.Read,
		})
	}
	return out
}

// MarkMessagesAsRead marks messages in an agent's inbox as read
func (t *Team) MarkMessagesAsRead(agentID string) {
	t.mutex.Lock()
	for i := range t.messages {
		if t.messages[i].To == agentID {
			t.messages[i].Read = true
		}
	}
	t.mutex.Unlock()
}
