package persistent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/registry"
)

// PersistentAgent wraps a core.Agent with TCP server for persistent communication
type PersistentAgent struct {
	ID         string              `json:"id"`
	Agent      *core.Agent         `json:"-"`
	Port       int                 `json:"port"`
	PID        int                 `json:"pid"`
	Server     *http.Server        `json:"-"`
	Status     registry.AgentStatus `json:"status"`
	StartedAt  time.Time           `json:"started_at"`
	LastSeen   time.Time           `json:"last_seen"`
	Role       string              `json:"role,omitempty"`
	mutex      sync.RWMutex
}

// PersistentTeam manages a collection of persistent agents
type PersistentTeam struct {
	parent      *core.Agent
	agents      map[string]*PersistentAgent
	registry    registry.AgentRegistry
	portRange   registry.PortRange
	mutex       sync.RWMutex
}

// DefaultPortRange returns the default port range for agents
func DefaultPortRange() registry.PortRange {
	return registry.PortRange{Start: 9000, End: 9099}
}

// NewPersistentTeam creates a new persistent team with the given parent agent and port range
func NewPersistentTeam(parent *core.Agent, portRange registry.PortRange) (*PersistentTeam, error) {
	// Create file-based registry for agent discovery
	reg, err := registry.NewFileRegistry(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	return &PersistentTeam{
		parent:    parent,
		agents:    make(map[string]*PersistentAgent),
		registry:  reg,
		portRange: portRange,
	}, nil
}

// SpawnAgent creates a new persistent agent with the given ID and role
func (pt *PersistentTeam) SpawnAgent(ctx context.Context, agentID, role string) (*PersistentAgent, error) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	// Check if agent already exists
	if existing, exists := pt.agents[agentID]; exists {
		if existing.Status == registry.StatusRunning {
			return existing, nil
		}
		// Clean up old agent
		pt.stopAgent(existing)
	}

	// Find available port
	port, err := pt.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("no available ports: %w", err)
	}

	// Create team context from existing converse system
	team, err := converse.NewTeamContext(pt.parent)
	if err != nil {
		return nil, fmt.Errorf("failed to create team context: %w", err)
	}

	// Add agent to team using existing system
	agent, _ := team.AddAgent(agentID)

	// Create persistent agent wrapper
	persistentAgent := &PersistentAgent{
		ID:        agentID,
		Agent:     agent,
		Port:      port,
		PID:       os.Getpid(), // For now, same process
		Status:    registry.StatusStarting,
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Role:      role,
	}

	// Start HTTP server for agent communication
	if err := pt.startAgentServer(persistentAgent); err != nil {
		return nil, fmt.Errorf("failed to start agent server: %w", err)
	}

	// Register with registry
	agentInfo := &registry.AgentInfo{
		ID:           agentID,
		Port:         port,
		PID:          os.Getpid(),
		Capabilities: []string{role}, // Use role as capability
		Endpoint:     fmt.Sprintf("localhost:%d", port),
		Status:       registry.StatusRunning,
		RegisteredAt: time.Now(),
		LastSeen:     time.Now(),
		Metadata:     map[string]string{"role": role, "spawned_by": "persistent_team"},
	}

	if err := pt.registry.RegisterAgent(ctx, agentInfo); err != nil {
		pt.stopAgent(persistentAgent)
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	pt.agents[agentID] = persistentAgent
	persistentAgent.Status = registry.StatusRunning

	return persistentAgent, nil
}

// GetAgent returns an existing persistent agent by ID
func (pt *PersistentTeam) GetAgent(agentID string) (*PersistentAgent, bool) {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	
	agent, exists := pt.agents[agentID]
	return agent, exists
}

// ListAgents returns all currently managed persistent agents
func (pt *PersistentTeam) ListAgents() []*PersistentAgent {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()

	agents := make([]*PersistentAgent, 0, len(pt.agents))
	for _, agent := range pt.agents {
		agents = append(agents, agent)
	}
	return agents
}

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

// StopAgent stops a persistent agent and cleans up resources
func (pt *PersistentTeam) StopAgent(ctx context.Context, agentID string) error {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	agent, exists := pt.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	pt.stopAgent(agent)
	delete(pt.agents, agentID)

	// Deregister from registry
	return pt.registry.DeregisterAgent(ctx, agentID)
}

// Close shuts down all persistent agents and cleans up resources
func (pt *PersistentTeam) Close() error {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	for _, agent := range pt.agents {
		pt.stopAgent(agent)
	}
	pt.agents = make(map[string]*PersistentAgent)

	return pt.registry.Close()
}

// findAvailablePort finds an available port in the configured range
func (pt *PersistentTeam) findAvailablePort() (int, error) {
	for port := pt.portRange.Start; port <= pt.portRange.End; port++ {
		if pt.isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", pt.portRange.Start, pt.portRange.End)
}

// isPortAvailable checks if a port is available for use
func (pt *PersistentTeam) isPortAvailable(port int) bool {
	// Check if any existing agent is using this port
	for _, agent := range pt.agents {
		if agent.Port == port {
			return false
		}
	}

	// Try to bind to the port to verify it's available
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	
	// Close the listener immediately
	listener.Close()
	return true
}

// startAgentServer starts the HTTP server for an agent
func (pt *PersistentTeam) startAgentServer(agent *PersistentAgent) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		agent.mutex.RLock()
		status := agent.Status
		agent.mutex.RUnlock()

		response := map[string]interface{}{
			"status":    status,
			"agent_id":  agent.ID,
			"uptime":    time.Since(agent.StartedAt).Seconds(),
			"last_seen": agent.LastSeen,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Message endpoint - now processes tasks through agent.Agent.Run()
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Update last seen
		agent.mutex.Lock()
		agent.LastSeen = time.Now()
		agent.Status = registry.StatusWorking
		agent.mutex.Unlock()

		// Parse request body
		var msgRequest struct {
			Input    string            `json:"input"`
			From     string            `json:"from,omitempty"`
			TaskID   string            `json:"task_id,omitempty"`
			Metadata map[string]string `json:"metadata,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&msgRequest); err != nil {
			agent.mutex.Lock()
			agent.Status = registry.StatusIdle
			agent.mutex.Unlock()
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Execute task through the actual agent
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()
		
		// Create team context for the agent (required for agent.Run())
		teamCtx := context.WithValue(ctx, "team", pt)
		
		result, err := agent.Agent.Run(teamCtx, msgRequest.Input)
		
		// Update status back to idle
		agent.mutex.Lock()
		agent.Status = registry.StatusIdle
		agent.mutex.Unlock()

		if err != nil {
			response := map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
				"agent_id": agent.ID,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"status":   "success",
			"result":   result,
			"agent_id": agent.ID,
			"task_id":  msgRequest.TaskID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", agent.Port),
		Handler: mux,
	}

	agent.Server = server

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Agent %s server error: %v\n", agent.ID, err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// stopAgent stops a single agent and cleans up its resources
func (pt *PersistentTeam) stopAgent(agent *PersistentAgent) {
	agent.mutex.Lock()
	defer agent.mutex.Unlock()

	agent.Status = registry.StatusStopping

	if agent.Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		agent.Server.Shutdown(ctx)
	}

	agent.Status = registry.StatusStopped
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

// NewPersistentTeamFromConfig creates a PersistentTeam from configuration
func NewPersistentTeamFromConfig(parent *core.Agent, cfg *config.PersistentAgentsConfig) (*PersistentTeam, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("persistent agents not enabled")
	}

	portStart := cfg.PortStart
	portEnd := cfg.PortEnd
	
	// Set default port range if not specified
	if portStart == 0 {
		portStart = 9001
	}
	if portEnd == 0 {
		portEnd = 9010
	}

	return NewPersistentTeam(parent, registry.PortRange{
		Start: portStart,
		End:   portEnd,
	})
}
