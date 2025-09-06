package team

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

// ListRoleNames returns all available role names
func (t *Team) ListRoleNames() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	names := make([]string, 0, len(t.roles))
	for name := range t.roles {
		names = append(names, name)
	}

	return names
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
