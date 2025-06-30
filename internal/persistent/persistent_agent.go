package persistent

import (
	"sync"
	"time"
	"net/http"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/registry"
	"github.com/marcodenic/agentry/internal/sessions"
)

// PersistentAgent wraps a core.Agent with TCP server for persistent communication
type PersistentAgent struct {
	ID             string                   `json:"id"`
	Agent          *core.Agent              `json:"-"`
	SessionAgent   *sessions.SessionAgent   `json:"-"`
	Port           int                      `json:"port"`
	PID            int                      `json:"pid"`
	Server         *http.Server             `json:"-"`
	Status         registry.AgentStatus     `json:"status"`
	StartedAt      time.Time                `json:"started_at"`
	LastSeen       time.Time                `json:"last_seen"`
	Role           string                   `json:"role,omitempty"`
	CurrentSession *sessions.SessionState   `json:"current_session,omitempty"`
	mutex          sync.RWMutex
}

// GetCurrentSessionInfo returns information about the current session
func (pa *PersistentAgent) GetCurrentSessionInfo() *sessions.SessionInfo {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()
	
	if pa.CurrentSession == nil {
		return nil
	}
	
	return &sessions.SessionInfo{
		ID:             pa.CurrentSession.ID,
		AgentID:        pa.CurrentSession.AgentID,
		Name:           pa.CurrentSession.Name,
		Description:    pa.CurrentSession.Description,
		CreatedAt:      pa.CurrentSession.CreatedAt,
		LastAccessedAt: pa.CurrentSession.LastAccessedAt,
		Status:         pa.CurrentSession.Status,
		WorkingDir:     pa.CurrentSession.WorkingDir,
	}
}

// UpdateLastSeen updates the agent's last seen timestamp
func (pa *PersistentAgent) UpdateLastSeen() {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	pa.LastSeen = time.Now()
}

// SetStatus updates the agent's status
func (pa *PersistentAgent) SetStatus(status registry.AgentStatus) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	pa.Status = status
}

// GetStatus returns the current agent status
func (pa *PersistentAgent) GetStatus() registry.AgentStatus {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()
	return pa.Status
}
