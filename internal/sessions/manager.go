package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/memory"
)

// SessionState represents the persistent state of an agent session
type SessionState struct {
	ID              string            `json:"id"`
	AgentID         string            `json:"agent_id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	CreatedAt       time.Time         `json:"created_at"`
	LastAccessedAt  time.Time         `json:"last_accessed_at"`
	WorkingDir      string            `json:"working_dir"`
	Prompt          string            `json:"prompt"`
	Vars            map[string]string `json:"vars"`
	Memory          []memory.Step     `json:"memory"`
	Status          SessionStatus     `json:"status"`
	Metadata        map[string]any    `json:"metadata"`
}

// SessionStatus represents the current status of a session
type SessionStatus string

const (
	StatusActive    SessionStatus = "active"
	StatusSuspended SessionStatus = "suspended"
	StatusTerminated SessionStatus = "terminated"
)

// SessionInfo provides basic information about a session
type SessionInfo struct {
	ID             string        `json:"id"`
	AgentID        string        `json:"agent_id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	CreatedAt      time.Time     `json:"created_at"`
	LastAccessedAt time.Time     `json:"last_accessed_at"`
	Status         SessionStatus `json:"status"`
	WorkingDir     string        `json:"working_dir"`
}

// SessionManager handles CRUD operations for agent sessions
type SessionManager interface {
	// CreateSession creates a new session for an agent
	CreateSession(ctx context.Context, req CreateSessionRequest) (*SessionState, error)
	
	// ListSessions returns all sessions, optionally filtered by agent ID
	ListSessions(ctx context.Context, agentID string) ([]*SessionInfo, error)
	
	// GetSession retrieves a specific session by ID
	GetSession(ctx context.Context, sessionID string) (*SessionState, error)
	
	// SaveSession saves the current state of a session
	SaveSession(ctx context.Context, state *SessionState) error
	
	// RestoreSession loads a session and updates its last accessed time
	RestoreSession(ctx context.Context, sessionID string) (*SessionState, error)
	
	// TerminateSession marks a session as terminated
	TerminateSession(ctx context.Context, sessionID string) error
	
	// CleanupOldSessions removes old sessions based on retention policy
	CleanupOldSessions(ctx context.Context, maxAge time.Duration) error
}

// CreateSessionRequest contains parameters for creating a new session
type CreateSessionRequest struct {
	AgentID     string            `json:"agent_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	WorkingDir  string            `json:"working_dir"`
	Prompt      string            `json:"prompt"`
	Vars        map[string]string `json:"vars"`
	Metadata    map[string]any    `json:"metadata"`
}

// FileSessionManager implements SessionManager using file-based storage
type FileSessionManager struct {
	baseDir string
}

// NewFileSessionManager creates a new file-based session manager
func NewFileSessionManager(baseDir string) (*FileSessionManager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}
	
	return &FileSessionManager{
		baseDir: baseDir,
	}, nil
}

// CreateSession creates a new session and saves it to disk
func (fsm *FileSessionManager) CreateSession(ctx context.Context, req CreateSessionRequest) (*SessionState, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	
	// Set default working directory if not provided
	workingDir := req.WorkingDir
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			workingDir = "/tmp"
		}
	}
	
	state := &SessionState{
		ID:              sessionID,
		AgentID:         req.AgentID,
		Name:            req.Name,
		Description:     req.Description,
		CreatedAt:       now,
		LastAccessedAt:  now,
		WorkingDir:      workingDir,
		Prompt:          req.Prompt,
		Vars:            req.Vars,
		Memory:          []memory.Step{},
		Status:          StatusActive,
		Metadata:        req.Metadata,
	}
	
	if state.Vars == nil {
		state.Vars = make(map[string]string)
	}
	if state.Metadata == nil {
		state.Metadata = make(map[string]any)
	}
	
	if err := fsm.SaveSession(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to save new session: %w", err)
	}
	
	return state, nil
}

// ListSessions returns all sessions, optionally filtered by agent ID
func (fsm *FileSessionManager) ListSessions(ctx context.Context, agentID string) ([]*SessionInfo, error) {
	files, err := filepath.Glob(filepath.Join(fsm.baseDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list session files: %w", err)
	}
	
	var sessions []*SessionInfo
	for _, file := range files {
		state, err := fsm.loadSessionFromFile(file)
		if err != nil {
			continue // Skip corrupted files
		}
		
		// Filter by agent ID if specified
		if agentID != "" && state.AgentID != agentID {
			continue
		}
		
		info := &SessionInfo{
			ID:             state.ID,
			AgentID:        state.AgentID,
			Name:           state.Name,
			Description:    state.Description,
			CreatedAt:      state.CreatedAt,
			LastAccessedAt: state.LastAccessedAt,
			Status:         state.Status,
			WorkingDir:     state.WorkingDir,
		}
		sessions = append(sessions, info)
	}
	
	return sessions, nil
}

// GetSession retrieves a specific session by ID
func (fsm *FileSessionManager) GetSession(ctx context.Context, sessionID string) (*SessionState, error) {
	return fsm.loadSessionFromFile(fsm.sessionFilePath(sessionID))
}

// SaveSession saves the current state of a session
func (fsm *FileSessionManager) SaveSession(ctx context.Context, state *SessionState) error {
	state.LastAccessedAt = time.Now()
	
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session state: %w", err)
	}
	
	filePath := fsm.sessionFilePath(state.ID)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	
	return nil
}

// RestoreSession loads a session and updates its last accessed time
func (fsm *FileSessionManager) RestoreSession(ctx context.Context, sessionID string) (*SessionState, error) {
	state, err := fsm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	
	// Update last accessed time and save
	state.LastAccessedAt = time.Now()
	if err := fsm.SaveSession(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to update session access time: %w", err)
	}
	
	return state, nil
}

// TerminateSession marks a session as terminated
func (fsm *FileSessionManager) TerminateSession(ctx context.Context, sessionID string) error {
	state, err := fsm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	
	state.Status = StatusTerminated
	return fsm.SaveSession(ctx, state)
}

// CleanupOldSessions removes old sessions based on retention policy
func (fsm *FileSessionManager) CleanupOldSessions(ctx context.Context, maxAge time.Duration) error {
	sessions, err := fsm.ListSessions(ctx, "")
	if err != nil {
		return err
	}
	
	cutoff := time.Now().Add(-maxAge)
	for _, session := range sessions {
		if session.LastAccessedAt.Before(cutoff) && session.Status != StatusActive {
			filePath := fsm.sessionFilePath(session.ID)
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove old session %s: %w", session.ID, err)
			}
		}
	}
	
	return nil
}

// sessionFilePath returns the file path for a session
func (fsm *FileSessionManager) sessionFilePath(sessionID string) string {
	return filepath.Join(fsm.baseDir, sessionID+".json")
}

// loadSessionFromFile loads a session from a file
func (fsm *FileSessionManager) loadSessionFromFile(filePath string) (*SessionState, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}
	
	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session state: %w", err)
	}
	
	return &state, nil
}
