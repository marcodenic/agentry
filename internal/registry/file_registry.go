package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileRegistry implements AgentRegistry using a JSON file for cross-platform persistence
type FileRegistry struct {
	agents     map[string]*AgentInfo
	health     map[string]*HealthMetrics
	mutex      sync.RWMutex
	configFile string
	events     []EventSubscriber
}

// RegistryConfig holds the configuration for the file registry
type RegistryConfig struct {
	ConfigDir string `json:"config_dir"`
	FileName  string `json:"file_name"`
}

// NewFileRegistry creates a new file-based agent registry
func NewFileRegistry(config *RegistryConfig) (*FileRegistry, error) {
	if config == nil {
		config = DefaultRegistryConfig()
	}

	// Ensure config directory exists
	if err := os.MkdirAll(config.ConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(config.ConfigDir, config.FileName)
	
	registry := &FileRegistry{
		agents:     make(map[string]*AgentInfo),
		health:     make(map[string]*HealthMetrics),
		configFile: configFile,
		events:     make([]EventSubscriber, 0),
	}

	// Load existing agents from file
	if err := registry.loadFromFile(); err != nil {
		// If file doesn't exist, that's ok - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load existing registry: %w", err)
		}
	}

	return registry, nil
}

// DefaultRegistryConfig returns the default configuration for the file registry
func DefaultRegistryConfig() *RegistryConfig {
	// Use cross-platform temporary directory
	tempDir := os.TempDir()
	agentryDir := filepath.Join(tempDir, "agentry")
	
	return &RegistryConfig{
		ConfigDir: agentryDir,
		FileName:  "agents.json",
	}
}

// RegisterAgent registers a new agent with the registry
func (r *FileRegistry) RegisterAgent(ctx context.Context, info *AgentInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if info.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Set registration time and initial status
	now := time.Now()
	info.RegisteredAt = now
	info.LastSeen = now
	if info.Status == "" {
		info.Status = StatusStarting
	}

	r.agents[info.ID] = info

	// Initialize health metrics
	r.health[info.ID] = &HealthMetrics{
		TasksCompleted: 0,
		TasksActive:    0,
		ErrorCount:     0,
		Uptime:         0,
	}

	if err := r.saveToFile(); err != nil {
		return fmt.Errorf("failed to persist agent registration: %w", err)
	}

	// Emit registration event
	r.emitEvent(&RegistryEvent{
		Type:      EventAgentRegistered,
		AgentID:   info.ID,
		Timestamp: now,
		Data: map[string]string{
			"endpoint": info.Endpoint,
			"role":     info.Role,
		},
	})

	return nil
}

// DeregisterAgent removes an agent from the registry
func (r *FileRegistry) DeregisterAgent(ctx context.Context, agentID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(r.agents, agentID)
	delete(r.health, agentID)

	if err := r.saveToFile(); err != nil {
		return fmt.Errorf("failed to persist agent deregistration: %w", err)
	}

	// Emit deregistration event
	r.emitEvent(&RegistryEvent{
		Type:      EventAgentDeregistered,
		AgentID:   agentID,
		Timestamp: time.Now(),
	})

	return nil
}

// UpdateAgent updates agent information
func (r *FileRegistry) UpdateAgent(ctx context.Context, agentID string, info *AgentInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Preserve registration time
	info.RegisteredAt = r.agents[agentID].RegisteredAt
	info.LastSeen = time.Now()
	r.agents[agentID] = info

	return r.saveToFile()
}

// GetAgent retrieves information about a specific agent
func (r *FileRegistry) GetAgent(ctx context.Context, agentID string) (*AgentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	// Return a copy to prevent external modification
	agentCopy := *agent
	return &agentCopy, nil
}

// ListAllAgents returns all registered agents
func (r *FileRegistry) ListAllAgents(ctx context.Context) ([]*AgentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	return agents, nil
}

// FindAgents finds agents with specific capabilities
func (r *FileRegistry) FindAgents(ctx context.Context, capabilities []string) ([]*AgentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var matchingAgents []*AgentInfo

	for _, agent := range r.agents {
		if r.hasCapabilities(agent, capabilities) {
			agentCopy := *agent
			matchingAgents = append(matchingAgents, &agentCopy)
		}
	}

	return matchingAgents, nil
}

// UpdateAgentStatus updates the status of an agent
func (r *FileRegistry) UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	oldStatus := agent.Status
	agent.Status = status
	agent.LastSeen = time.Now()

	if err := r.saveToFile(); err != nil {
		return err
	}

	// Emit status change event
	if oldStatus != status {
		r.emitEvent(&RegistryEvent{
			Type:      EventAgentStatusChange,
			AgentID:   agentID,
			Timestamp: time.Now(),
			Data: map[string]string{
				"old_status": string(oldStatus),
				"new_status": string(status),
			},
		})
	}

	return nil
}

// UpdateAgentHealth updates health metrics for an agent
func (r *FileRegistry) UpdateAgentHealth(ctx context.Context, agentID string, health *HealthMetrics) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	r.health[agentID] = health
	return nil
}

// GetAgentHealth retrieves health metrics for an agent
func (r *FileRegistry) GetAgentHealth(ctx context.Context, agentID string) (*HealthMetrics, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	health, exists := r.health[agentID]
	if !exists {
		return nil, fmt.Errorf("health metrics for agent %s not found", agentID)
	}

	// Return a copy
	healthCopy := *health
	return &healthCopy, nil
}

// Heartbeat updates the last seen time for an agent
func (r *FileRegistry) Heartbeat(ctx context.Context, agentID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.LastSeen = time.Now()
	
	// Update status to running if it was starting
	if agent.Status == StatusStarting {
		agent.Status = StatusIdle
	}

	return r.saveToFile()
}

// Close closes the registry and any underlying resources
func (r *FileRegistry) Close() error {
	// Save final state
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.saveToFile()
}

// AddEventSubscriber adds an event subscriber
func (r *FileRegistry) AddEventSubscriber(subscriber EventSubscriber) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.events = append(r.events, subscriber)
}

// loadFromFile loads agents from the JSON file
func (r *FileRegistry) loadFromFile() error {
	data, err := os.ReadFile(r.configFile)
	if err != nil {
		return err
	}

	var fileData struct {
		Agents map[string]*AgentInfo    `json:"agents"`
		Health map[string]*HealthMetrics `json:"health"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}

	if fileData.Agents != nil {
		r.agents = fileData.Agents
	}
	if fileData.Health != nil {
		r.health = fileData.Health
	}

	return nil
}

// saveToFile saves agents to the JSON file
func (r *FileRegistry) saveToFile() error {
	fileData := struct {
		Agents map[string]*AgentInfo    `json:"agents"`
		Health map[string]*HealthMetrics `json:"health"`
	}{
		Agents: r.agents,
		Health: r.health,
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}

	// Write to temporary file first, then rename for atomic update
	tempFile := r.configFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, r.configFile); err != nil {
		os.Remove(tempFile) // Clean up on failure
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// hasCapabilities checks if an agent has all required capabilities
func (r *FileRegistry) hasCapabilities(agent *AgentInfo, required []string) bool {
	agentCaps := make(map[string]bool)
	for _, cap := range agent.Capabilities {
		agentCaps[cap] = true
	}

	for _, reqCap := range required {
		if !agentCaps[reqCap] {
			return false
		}
	}

	return true
}

// emitEvent sends an event to all subscribers
func (r *FileRegistry) emitEvent(event *RegistryEvent) {
	for _, subscriber := range r.events {
		// Send events asynchronously to avoid blocking
		go func(sub EventSubscriber, evt *RegistryEvent) {
			if err := sub.OnEvent(evt); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Event subscriber error: %v\n", err)
			}
		}(subscriber, event)
	}
}
