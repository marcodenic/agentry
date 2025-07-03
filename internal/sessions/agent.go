package sessions

import (
	"context"
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/core"
)

// SessionAgent wraps a core.Agent with session management capabilities
type SessionAgent struct {
	*core.Agent
	sessionManager SessionManager
	currentSession *SessionState
}

// NewSessionAgent creates a new session-aware agent
func NewSessionAgent(agent *core.Agent, sessionManager SessionManager) *SessionAgent {
	return &SessionAgent{
		Agent:          agent,
		sessionManager: sessionManager,
	}
}

// CreateSession creates a new session for this agent
func (sa *SessionAgent) CreateSession(ctx context.Context, name, description string) (*SessionState, error) {
	req := CreateSessionRequest{
		AgentID:     sa.ID.String(),
		Name:        name,
		Description: description,
		Prompt:      sa.Prompt,
		Vars:        sa.Vars,
	}
	
	// Get current working directory
	if wd, err := os.Getwd(); err == nil {
		req.WorkingDir = wd
	}
	
	session, err := sa.sessionManager.CreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	sa.currentSession = session
	return session, nil
}

// LoadSession loads an existing session and restores agent state
func (sa *SessionAgent) LoadSession(ctx context.Context, sessionID string) error {
	session, err := sa.sessionManager.RestoreSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to restore session: %w", err)
	}
	
	// Restore agent state from session
	sa.currentSession = session
	sa.Prompt = session.Prompt
	sa.Vars = session.Vars
	
	// Restore memory
	if len(session.Memory) > 0 {
		sa.Mem.SetHistory(session.Memory)
	}
	
	// Change working directory if specified
	if session.WorkingDir != "" {
		if err := os.Chdir(session.WorkingDir); err != nil {
			// Don't fail if we can't change directory, just log it
			fmt.Printf("Warning: failed to change to working directory %s: %v\n", session.WorkingDir, err)
		}
	}
	
	return nil
}

// SaveSession saves the current agent state to the session
func (sa *SessionAgent) SaveSession(ctx context.Context) error {
	if sa.currentSession == nil {
		return fmt.Errorf("no active session")
	}
	
	// Update session state with current agent state
	sa.currentSession.Prompt = sa.Prompt
	sa.currentSession.Vars = sa.Vars
	sa.currentSession.Memory = sa.Mem.History()
	
	// Update working directory
	if wd, err := os.Getwd(); err == nil {
		sa.currentSession.WorkingDir = wd
	}
	
	return sa.sessionManager.SaveSession(ctx, sa.currentSession)
}

// RunWithSession runs the agent with automatic session saving
func (sa *SessionAgent) RunWithSession(ctx context.Context, input string) (string, error) {
	// Run the normal agent logic
	result, err := sa.Agent.Run(ctx, input)
	
	// Save session state after execution
	if sa.currentSession != nil {
		if saveErr := sa.SaveSession(ctx); saveErr != nil {
			fmt.Printf("Warning: failed to save session: %v\n", saveErr)
		}
	}
	
	return result, err
}

// GetCurrentSession returns the current session state
func (sa *SessionAgent) GetCurrentSession() *SessionState {
	return sa.currentSession
}

// ListSessions returns all sessions for this agent
func (sa *SessionAgent) ListSessions(ctx context.Context) ([]*SessionInfo, error) {
	return sa.sessionManager.ListSessions(ctx, sa.ID.String())
}

// TerminateCurrentSession terminates the current session
func (sa *SessionAgent) TerminateCurrentSession(ctx context.Context) error {
	if sa.currentSession == nil {
		return fmt.Errorf("no active session")
	}
	
	err := sa.sessionManager.TerminateSession(ctx, sa.currentSession.ID)
	if err == nil {
		sa.currentSession = nil
	}
	return err
}

// SuspendSession suspends the current session (saves state and marks as suspended)
func (sa *SessionAgent) SuspendSession(ctx context.Context) error {
	if sa.currentSession == nil {
		return fmt.Errorf("no active session")
	}
	
	sa.currentSession.Status = StatusSuspended
	return sa.SaveSession(ctx)
}

// ResumeSession resumes a suspended session
func (sa *SessionAgent) ResumeSession(ctx context.Context, sessionID string) error {
	if err := sa.LoadSession(ctx, sessionID); err != nil {
		return err
	}
	
	if sa.currentSession.Status != StatusSuspended {
		return fmt.Errorf("session %s is not suspended", sessionID)
	}
	
	sa.currentSession.Status = StatusActive
	return sa.SaveSession(ctx)
}

// SessionAgentFactory creates session-aware agents
type SessionAgentFactory struct {
	sessionManager SessionManager
}

// NewSessionAgentFactory creates a new factory for session agents
func NewSessionAgentFactory(sessionManager SessionManager) *SessionAgentFactory {
	return &SessionAgentFactory{
		sessionManager: sessionManager,
	}
}

// WrapAgent wraps a core agent with session management
func (f *SessionAgentFactory) WrapAgent(agent *core.Agent) *SessionAgent {
	return NewSessionAgent(agent, f.sessionManager)
}

// CreateAgentWithSession creates a new agent and immediately creates a session for it
func (f *SessionAgentFactory) CreateAgentWithSession(ctx context.Context, agent *core.Agent, sessionName, description string) (*SessionAgent, error) {
	sessionAgent := f.WrapAgent(agent)
	
	_, err := sessionAgent.CreateSession(ctx, sessionName, description)
	if err != nil {
		return nil, err
	}
	
	return sessionAgent, nil
}
