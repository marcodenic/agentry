package registry

// This file has been refactored for maintainability.
// The original implementation has been split into focused modules:
// - inmemory_registry_core.go: core InMemoryRegistry struct and constructor
// - inmemory_registry_registration.go: agent registration and deregistration operations
// - inmemory_registry_queries.go: agent query operations (get, list, find)
// - inmemory_registry_health.go: status and health management operations
// - inmemory_registry_events.go: event subscription, notification, and cleanup functionality
