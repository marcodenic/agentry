package persistent

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/sessions"
)

// Session Management Methods for PersistentAgent

// CreateSession creates a new session for this persistent agent
func (pa *PersistentAgent) CreateSession(ctx context.Context, name, description string) (*sessions.SessionState, error) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if pa.SessionAgent == nil {
		return nil, fmt.Errorf("session agent not initialized")
	}
	
	session, err := pa.SessionAgent.CreateSession(ctx, name, description)
	if err != nil {
		return nil, err
	}
	
	pa.CurrentSession = session
	return session, nil
}

// LoadSession loads an existing session for this persistent agent
func (pa *PersistentAgent) LoadSession(ctx context.Context, sessionID string) error {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if pa.SessionAgent == nil {
		return fmt.Errorf("session agent not initialized")
	}
	
	err := pa.SessionAgent.LoadSession(ctx, sessionID)
	if err != nil {
		return err
	}
	
	pa.CurrentSession = pa.SessionAgent.GetCurrentSession()
	return nil
}

// SaveCurrentSession saves the current session state
func (pa *PersistentAgent) SaveCurrentSession(ctx context.Context) error {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()
	
	if pa.SessionAgent == nil {
		return fmt.Errorf("session agent not initialized")
	}
	
	return pa.SessionAgent.SaveSession(ctx)
}

// ListSessions returns all sessions for this agent
func (pa *PersistentAgent) ListSessions(ctx context.Context) ([]*sessions.SessionInfo, error) {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()
	
	if pa.SessionAgent == nil {
		return nil, fmt.Errorf("session agent not initialized")
	}
	
	return pa.SessionAgent.ListSessions(ctx)
}

// TerminateCurrentSession terminates the current session
func (pa *PersistentAgent) TerminateCurrentSession(ctx context.Context) error {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if pa.SessionAgent == nil {
		return fmt.Errorf("session agent not initialized")
	}
	
	err := pa.SessionAgent.TerminateCurrentSession(ctx)
	if err == nil {
		pa.CurrentSession = nil
	}
	return err
}

// SuspendCurrentSession suspends the current session
func (pa *PersistentAgent) SuspendCurrentSession(ctx context.Context) error {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if pa.SessionAgent == nil {
		return fmt.Errorf("session agent not initialized")
	}
	
	err := pa.SessionAgent.SuspendSession(ctx)
	if err == nil {
		pa.CurrentSession = pa.SessionAgent.GetCurrentSession()
	}
	return err
}

// ResumeSession resumes a suspended session
func (pa *PersistentAgent) ResumeSession(ctx context.Context, sessionID string) error {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if pa.SessionAgent == nil {
		return fmt.Errorf("session agent not initialized")
	}
	
	err := pa.SessionAgent.ResumeSession(ctx, sessionID)
	if err == nil {
		pa.CurrentSession = pa.SessionAgent.GetCurrentSession()
	}
	return err
}
