package registry

// This file has been refactored for maintainability.
// The original implementation has been split into focused modules:
// - discovery_core.go: core DiscoveryService struct, types, and constructor
// - discovery_search.go: main discovery logic (FindBestAgents, FindAvailableAgent)
// - discovery_queries.go: query functions (GetAgentCapabilities, GetAgentsByRole, GetClusterStatus)
// - discovery_scoring.go: scoring and sorting algorithms for agent selection
// - discovery_utils.go: utility functions for filtering and validation
