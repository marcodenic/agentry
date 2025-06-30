package core

// This file has been refactored for maintainability.
// The original implementation has been split into focused modules:
// - orchestrator_core.go: core types (TeamOrchestrator, AgentStatus, TeamMessage) and constructor
// - orchestrator_agents.go: agent management (RegisterAgent, UpdateAgentStatus, GetAvailableAgents)  
// - orchestrator_tasks.go: task management (AssignTask, CompleteTask)
// - orchestrator_messaging.go: messaging system (SendMessage, GetMessages)
// - orchestrator_status.go: status monitoring and reporting (GetTeamStatus, GetSystemPrompt, formatTeamStatus)
// - orchestrator_coordination.go: coordination logic (CoordinateTask, ProcessTeamCommand)
