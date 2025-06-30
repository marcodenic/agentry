package persistent

// This file has been refactored for maintainability.
// The original implementation has been split into focused modules:
// - persistent_team_core.go: core PersistentTeam struct, constructors, and lifecycle management
// - persistent_team_agents.go: agent management (SpawnAgent, GetAgent, ListAgents, StopAgent)
// - persistent_team_communication.go: communication between agents (SendMessage, Call interface)
// - persistent_team_utils.go: utility functions (port management, agent server lifecycle)
