package registry

// This file has been refactored for maintainability.
// The original implementation has been split into focused modules:
// - file_registry_core.go: core FileRegistry struct, config, and constructor
// - file_registry_registration.go: agent registration and deregistration operations  
// - file_registry_queries.go: agent query operations (get, list, find)
// - file_registry_health.go: status and health management operations
// - file_registry_persistence.go: file I/O operations for JSON persistence
// - file_registry_events.go: event subscription and emission functionality
