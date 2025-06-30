package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/marcodenic/agentry/internal/sessions"
)

// Session Management Methods for PersistentTeam

// CreateAgentSession creates a new session for a specific agent
func (pt *PersistentTeam) CreateAgentSession(ctx context.Context, agentID, sessionName, description string) (*sessions.SessionState, error) {
	agent, exists := pt.GetAgent(agentID)
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}
	
	return agent.CreateSession(ctx, sessionName, description)
}

// LoadAgentSession loads a session for a specific agent
func (pt *PersistentTeam) LoadAgentSession(ctx context.Context, agentID, sessionID string) error {
	agent, exists := pt.GetAgent(agentID)
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}
	
	return agent.LoadSession(ctx, sessionID)
}

// ListAgentSessions lists all sessions for a specific agent
func (pt *PersistentTeam) ListAgentSessions(ctx context.Context, agentID string) ([]*sessions.SessionInfo, error) {
	agent, exists := pt.GetAgent(agentID)
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}
	
	return agent.ListSessions(ctx)
}

// ListAllSessions lists all sessions across all agents
func (pt *PersistentTeam) ListAllSessions(ctx context.Context) ([]*sessions.SessionInfo, error) {
	return pt.sessionManager.ListSessions(ctx, "")
}

// CleanupOldSessions removes old sessions based on retention policy
func (pt *PersistentTeam) CleanupOldSessions(ctx context.Context, maxAge time.Duration) error {
	return pt.sessionManager.CleanupOldSessions(ctx, maxAge)
}
