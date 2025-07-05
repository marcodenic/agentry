package team

import (
	"fmt"
	"net"
)

// findAvailablePort finds an available port in the configured range
func (t *Team) findAvailablePort() (int, error) {
	for port := t.portRange.Start; port <= t.portRange.End; port++ {
		if t.isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", t.portRange.Start, t.portRange.End)
}

// isPortAvailable checks if a port is available for use
func (t *Team) isPortAvailable(port int) bool {
	// Check if any existing agent is using this port
	for _, agent := range t.agents {
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

// AddRole adds a role configuration to the team
func (t *Team) AddRole(role *RoleConfig) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.roles[role.Name] = role
}

// GetRole returns a role configuration by name
func (t *Team) GetRole(name string) *RoleConfig {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.roles[name]
}

// ListRoles returns all available roles
func (t *Team) ListRoles() []*RoleConfig {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	roles := make([]*RoleConfig, 0, len(t.roles))
	for _, role := range t.roles {
		roles = append(roles, role)
	}

	return roles
}

// SetMaxTurns sets the maximum number of turns for conversations
func (t *Team) SetMaxTurns(maxTurns int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.maxTurns = maxTurns
}

// GetMaxTurns returns the maximum number of turns
func (t *Team) GetMaxTurns() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.maxTurns
}

// GetName returns the team name
func (t *Team) GetName() string {
	return t.name
}

// GetParent returns the parent agent
func (t *Team) GetParent() interface{} {
	return t.parent
}
