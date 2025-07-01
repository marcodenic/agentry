package collaboration

import (
	"fmt"
	"sync"
	"time"
)

// CollaborationEngine manages true multi-agent collaboration
type CollaborationEngine struct {
	team        *Team
	eventBus    *EventBus
	workflow    *WorkflowOrchestrator
	fileManager *CollaborativeFileManager
	statusBoard *StatusBoard
	mutex       sync.RWMutex
}

// EventBus handles real-time agent-to-agent communication
type EventBus struct {
	subscribers map[string][]chan AgentEvent
	mutex       sync.RWMutex
}

type AgentEvent struct {
	Type      string                 `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"` // empty for broadcast
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	MessageID string                 `json:"message_id"`
	ReplyTo   string                 `json:"reply_to,omitempty"`
	Priority  string                 `json:"priority"` // high, normal, low
}

// WorkflowOrchestrator manages dependencies and execution order
type WorkflowOrchestrator struct {
	workflows map[string]*Workflow
	running   map[string]*WorkflowExecution
	mutex     sync.RWMutex
}

type Workflow struct {
	ID           string                       `json:"id"`
	Name         string                       `json:"name"`
	Steps        []WorkflowStep               `json:"steps"`
	Dependencies map[string][]string          `json:"dependencies"` // step_id -> list of prerequisite step_ids
	Conditions   map[string]WorkflowCondition `json:"conditions"`
}

type WorkflowStep struct {
	ID         string                 `json:"id"`
	AgentID    string                 `json:"agent_id"`
	Task       string                 `json:"task"`
	Timeout    time.Duration          `json:"timeout"`
	Parameters map[string]interface{} `json:"parameters"`
	OnSuccess  []string               `json:"on_success"` // next steps to execute
	OnFailure  []string               `json:"on_failure"` // steps to execute on failure
	RetryCount int                    `json:"retry_count"`
}

type WorkflowCondition struct {
	Type       string                 `json:"type"` // file_exists, agent_status, custom
	Parameters map[string]interface{} `json:"parameters"`
}

type WorkflowExecution struct {
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"`      // running, completed, failed, paused
	StepStatus  map[string]string      `json:"step_status"` // step_id -> status
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Results     map[string]interface{} `json:"results"`
}

// CollaborativeFileManager handles file locking and change notifications
type CollaborativeFileManager struct {
	locks         map[string]*FileLock
	watchers      map[string][]string // file_path -> list of agent_ids watching
	changeHistory []FileChange
	mutex         sync.RWMutex
}

type FileLock struct {
	FilePath   string    `json:"file_path"`
	AgentID    string    `json:"agent_id"`
	LockType   string    `json:"lock_type"` // read, write, exclusive
	AcquiredAt time.Time `json:"acquired_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type FileChange struct {
	FilePath   string                 `json:"file_path"`
	AgentID    string                 `json:"agent_id"`
	ChangeType string                 `json:"change_type"` // created, modified, deleted, moved
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// StatusBoard provides real-time status visibility for all agents
type StatusBoard struct {
	agentStatus   map[string]*AgentStatus
	globalStatus  *GlobalStatus
	statusHistory []StatusUpdate
	mutex         sync.RWMutex
}

type AgentStatus struct {
	AgentID      string                 `json:"agent_id"`
	Status       string                 `json:"status"` // idle, working, waiting, blocked, error
	CurrentTask  string                 `json:"current_task"`
	Progress     float64                `json:"progress"` // 0.0 to 1.0
	ETA          *time.Time             `json:"eta"`
	Dependencies []string               `json:"dependencies"` // what this agent is waiting for
	Blocking     []string               `json:"blocking"`     // what other agents are waiting for this agent
	Metadata     map[string]interface{} `json:"metadata"`
	LastUpdate   time.Time              `json:"last_update"`
}

type GlobalStatus struct {
	OverallProgress float64                `json:"overall_progress"`
	ActiveAgents    int                    `json:"active_agents"`
	CompletedTasks  int                    `json:"completed_tasks"`
	PendingTasks    int                    `json:"pending_tasks"`
	Bottlenecks     []string               `json:"bottlenecks"` // agent_ids that are blocking others
	Metrics         map[string]interface{} `json:"metrics"`
	LastUpdate      time.Time              `json:"last_update"`
}

type StatusUpdate struct {
	AgentID   string                 `json:"agent_id"`
	OldStatus string                 `json:"old_status"`
	NewStatus string                 `json:"new_status"`
	Reason    string                 `json:"reason"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewCollaborationEngine creates a new collaboration engine
func NewCollaborationEngine(team *Team) *CollaborationEngine {
	return &CollaborationEngine{
		team:        team,
		eventBus:    NewEventBus(),
		workflow:    NewWorkflowOrchestrator(),
		fileManager: NewCollaborativeFileManager(),
		statusBoard: NewStatusBoard(),
	}
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan AgentEvent),
	}
}

// NewWorkflowOrchestrator creates a new workflow orchestrator
func NewWorkflowOrchestrator() *WorkflowOrchestrator {
	return &WorkflowOrchestrator{
		workflows: make(map[string]*Workflow),
		running:   make(map[string]*WorkflowExecution),
	}
}

// NewCollaborativeFileManager creates a new collaborative file manager
func NewCollaborativeFileManager() *CollaborativeFileManager {
	return &CollaborativeFileManager{
		locks:         make(map[string]*FileLock),
		watchers:      make(map[string][]string),
		changeHistory: make([]FileChange, 0),
	}
}

// NewStatusBoard creates a new status board
func NewStatusBoard() *StatusBoard {
	return &StatusBoard{
		agentStatus:   make(map[string]*AgentStatus),
		globalStatus:  &GlobalStatus{},
		statusHistory: make([]StatusUpdate, 0),
	}
}

// Event Bus Methods

// Subscribe allows an agent to subscribe to events
func (eb *EventBus) Subscribe(agentID string, eventTypes []string) chan AgentEvent {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	ch := make(chan AgentEvent, 100) // Buffered channel to prevent blocking

	for _, eventType := range eventTypes {
		if eb.subscribers[eventType] == nil {
			eb.subscribers[eventType] = make([]chan AgentEvent, 0)
		}
		eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	}

	return ch
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(event AgentEvent) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	event.Timestamp = time.Now()
	if event.MessageID == "" {
		event.MessageID = fmt.Sprintf("%s_%d", event.From, time.Now().UnixNano())
	}

	// Send to specific agent if specified
	if event.To != "" {
		if subscribers, ok := eb.subscribers[event.Type]; ok {
			for _, ch := range subscribers {
				select {
				case ch <- event:
				default:
					// Channel is full, skip this subscriber
				}
			}
		}
	} else {
		// Broadcast to all subscribers of this event type
		if subscribers, ok := eb.subscribers[event.Type]; ok {
			for _, ch := range subscribers {
				select {
				case ch <- event:
				default:
					// Channel is full, skip this subscriber
				}
			}
		}
	}
}

// File Manager Methods

// AcquireLock attempts to acquire a file lock
func (fm *CollaborativeFileManager) AcquireLock(filePath, agentID, lockType string, duration time.Duration) (*FileLock, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// Check if file is already locked
	if existingLock, exists := fm.locks[filePath]; exists {
		if existingLock.ExpiresAt.After(time.Now()) {
			if existingLock.LockType == "exclusive" || lockType == "exclusive" {
				return nil, fmt.Errorf("file %s is locked by agent %s", filePath, existingLock.AgentID)
			}
			if existingLock.LockType == "write" && lockType == "write" {
				return nil, fmt.Errorf("file %s has a write lock by agent %s", filePath, existingLock.AgentID)
			}
		}
	}

	lock := &FileLock{
		FilePath:   filePath,
		AgentID:    agentID,
		LockType:   lockType,
		AcquiredAt: time.Now(),
		ExpiresAt:  time.Now().Add(duration),
	}

	fm.locks[filePath] = lock
	return lock, nil
}

// ReleaseLock releases a file lock
func (fm *CollaborativeFileManager) ReleaseLock(filePath, agentID string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if lock, exists := fm.locks[filePath]; exists {
		if lock.AgentID == agentID {
			delete(fm.locks, filePath)
			return nil
		}
		return fmt.Errorf("lock owned by different agent")
	}
	return fmt.Errorf("no lock found for file %s", filePath)
}

// NotifyFileChange notifies watchers of file changes
func (fm *CollaborativeFileManager) NotifyFileChange(filePath, agentID, changeType string, metadata map[string]interface{}) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	change := FileChange{
		FilePath:   filePath,
		AgentID:    agentID,
		ChangeType: changeType,
		Timestamp:  time.Now(),
		Metadata:   metadata,
	}

	fm.changeHistory = append(fm.changeHistory, change)

	// Notify watchers (this would integrate with the event bus)
	if watchers, exists := fm.watchers[filePath]; exists {
		for _, watcherID := range watchers {
			// Send file change event to watcher
			// This would use the event bus
			_ = watcherID // placeholder
		}
	}
}

// Status Board Methods

// UpdateAgentStatus updates an agent's status
func (sb *StatusBoard) UpdateAgentStatus(agentID string, status *AgentStatus) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	oldStatus := ""
	if existing, exists := sb.agentStatus[agentID]; exists {
		oldStatus = existing.Status
	}

	status.LastUpdate = time.Now()
	sb.agentStatus[agentID] = status

	// Record status change
	update := StatusUpdate{
		AgentID:   agentID,
		OldStatus: oldStatus,
		NewStatus: status.Status,
		Timestamp: time.Now(),
	}
	sb.statusHistory = append(sb.statusHistory, update)

	// Update global status
	sb.updateGlobalStatus()
}

// updateGlobalStatus recalculates global metrics
func (sb *StatusBoard) updateGlobalStatus() {
	activeAgents := 0
	totalProgress := 0.0
	var bottlenecks []string

	for agentID, status := range sb.agentStatus {
		if status.Status == "working" {
			activeAgents++
		}
		totalProgress += status.Progress

		if len(status.Blocking) > 0 {
			bottlenecks = append(bottlenecks, agentID)
		}
	}

	sb.globalStatus.ActiveAgents = activeAgents
	sb.globalStatus.OverallProgress = totalProgress / float64(len(sb.agentStatus))
	sb.globalStatus.Bottlenecks = bottlenecks
	sb.globalStatus.LastUpdate = time.Now()
}

// GetGlobalStatus returns the current global status
func (sb *StatusBoard) GetGlobalStatus() *GlobalStatus {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.globalStatus
}

// GetAgentStatus returns an agent's current status
func (sb *StatusBoard) GetAgentStatus(agentID string) *AgentStatus {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.agentStatus[agentID]
}
