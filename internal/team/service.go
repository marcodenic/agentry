package team

import (
    "context"
    "sort"
)

// GetAgents returns a list of all agent names in the team.
func (t *Team) GetAgents() []string {
    t.mutex.RLock()
    defer t.mutex.RUnlock()
    out := make([]string, 0, len(t.agentsByName))
    for name := range t.agentsByName { out = append(out, name) }
    sort.Strings(out)
    return out
}

// Names returns a list of all agent names in the team.
func (t *Team) Names() []string { return t.GetAgents() }

// ===== contracts.TeamService Implementation =====

// SpawnedAgentNames returns currently running agent instances
func (t *Team) SpawnedAgentNames() []string { return t.GetAgents() }

// AvailableRoleNames returns role names from configuration files
func (t *Team) AvailableRoleNames() []string { return t.ListRoleNames() }

// DelegateTask delegates a task to a role (spawning if needed)
func (t *Team) DelegateTask(ctx context.Context, role, task string) (string, error) { return t.Call(ctx, role, task) }

// GetInbox returns an agent's inbox messages
func (t *Team) GetInbox(agentID string) []map[string]interface{} { return t.GetAgentInbox(agentID) }

// MarkInboxRead marks an agent's messages as read
func (t *Team) MarkInboxRead(agentID string) { t.MarkMessagesAsRead(agentID) }

// GetCoordinationHistory returns coordination event history
func (t *Team) GetCoordinationHistory(limit int) []string { return t.CoordinationHistoryStrings(limit) }

// GetTeamAgents returns a list of all team agents with role information.
func (t *Team) GetTeamAgents() []*Agent { return t.ListAgents() }

