package hubgrpc

// This file has been refactored for maintainability.
// The original implementation has been split into focused modules:
// - hub_core.go: core Server struct, constructor, and basic infrastructure
// - hub_agents.go: agent operations (Spawn, SendMessage, Trace)
// - hub_registry.go: registry operations (RegisterAgent, DeregisterAgent, GetAgent, ListAgents)
// - hub_health.go: health and status management (UpdateAgentStatus, UpdateHealth, Heartbeat)
// - hub_discovery.go: discovery operations (FindAgents, GetClusterStatus)
// - hub_utils.go: utility functions for data conversion and filtering
