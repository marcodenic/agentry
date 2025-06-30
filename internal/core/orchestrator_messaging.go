package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/debug"
)

// SendMessage sends a message between team agents
func (to *TeamOrchestrator) SendMessage(from, targetAgent, msgType, content string, data map[string]interface{}) {
	to.mutex.Lock()
	defer to.mutex.Unlock()
	
	message := TeamMessage{
		ID:        uuid.New(),
		From:      from,
		To:        targetAgent,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
		Data:      data,
	}
	
	to.messageQueue = append(to.messageQueue, message)
	
	debug.Printf("TeamOrchestrator: Message from %s to %s: %s", from, targetAgent, content)
}

// GetMessages retrieves messages for a specific agent
func (to *TeamOrchestrator) GetMessages(agentName string) []TeamMessage {
	to.mutex.RLock()
	defer to.mutex.RUnlock()
	
	var messages []TeamMessage
	for _, msg := range to.messageQueue {
		if msg.To == agentName || msg.To == "" { // Direct messages or broadcasts
			messages = append(messages, msg)
		}
	}
	
	return messages
}
