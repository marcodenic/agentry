# Agentry Code Consolidation Analysis

## Problem Statement
Agent0 is not aware of available agent roles from YAML configs because there are **multiple overlapping systems** for role/prompt/injection/team creation that have accumulated over development. This creates confusion and makes it impossible to distinguish between "available roles" and "currently spawned agents".

## Current State: Multiple Overlapping Systems

### 1. **Agent Creation Systems** (4+ ways to create agents)
- `Team.Add(name, *core.Agent)` - adds existing agent
- `Team.AddAgent(name)` - creates new agent with generic prompt
- `Team.SpawnAgent(ctx, name, role)` - creates agent with role config
- `Team.Call()` auto-spawning - creates agents on-demand

### 2. **Role Configuration Systems** (3+ ways)
- `LoadRolesFromIncludePaths()` - loads from YAML files in team package
- `core.GetDefaultPrompt()` - loads agent_0.yaml specifically 
- Hardcoded fallback prompts in multiple places
- Role configs stored in `Team.roles` map but not exposed to Agent0

### 3. **Prompt/Context Injection Systems** (4+ ways)
- `buildContextMinimal()` in team.go with hardcoded project context
- Role-specific prompts from YAML files
- Inbox injection in `Team.Call()`
- Various hardcoded context strings scattered throughout

### 4. **Tool Registration Systems** (3+ ways)
- `tool.DefaultRegistry()` - base tools
- `Team.RegisterAgentTool()` - adds "agent" tool
- Team builtins added separately via teamAPI interface
- Ad-hoc tool filtering (removing "agent" tool to prevent cascading)

### 5. **Model Configuration Systems** (2+ ways)
- Config-based model selection in `buildAgent()`
- Role-specific model configs in YAML files (partially implemented)

## Root Cause: Missing Interface Between Team.roles and Agent0

The **core issue** is that while `Team.roles` contains all available roles from YAML configs:
- ✅ `Team.roles` is populated by `LoadRolesFromIncludePaths()`
- ✅ `Team.ListRoleNames()` returns available role names
- ❌ The `teamAPI` interface only has `Names()` (spawned agents) and `ListRoles()` but...
- ❌ **`Team.ListRoles()` is NOT implemented for the `teamAPI` interface**

Look at the interface mismatch:

```go
// internal/tool/builtins_team.go - teamAPI interface
type teamAPI interface {
    Names() []string      // ✅ implemented - returns spawned agents
    ListRoles() []string  // ❌ NOT implemented - should return available roles
}

// internal/team/utils.go - Team methods
func (t *Team) ListRoles() []*RoleConfig      // ✅ exists but returns wrong type
func (t *Team) ListRoleNames() []string       // ✅ exists and returns []string
```

## Impact Analysis

### What Works ✅
- Role loading from YAML files
- Agent spawning with role configs
- Team coordination and messaging
- Basic context injection

### What's Broken ❌
- Agent0 can't see available roles (only spawned agents)
- `team_status` and `check_agent` tools only show empty lists at startup
- Multiple context injection systems conflict
- Inconsistent tool registry handling
- Model configs in role YAML not fully utilized

## Consolidation Plan

### Phase 1: Fix Agent0 Role Awareness (Immediate)
1. **Fix teamAPI Interface Mismatch**
   ```go
   // Add missing method to Team struct
   func (t *Team) ListRoles() []string {
       return t.ListRoleNames()
   }
   ```

2. **Add Available Roles Tool**
   ```go
   "available_roles": {
       Desc: "List all available agent roles from configuration",
       Exec: func(ctx context.Context, args map[string]any) (string, error) {
           t := ctx.Value(TeamContextKey).(teamAPI)
           roles := t.ListRoles()
           // Format as helpful response
       }
   }
   ```

3. **Update team_status to Show Both**
   - Available roles from config
   - Currently spawned agents
   - Distinction between the two

### Phase 2: Consolidate Agent Creation (Next)
1. **Single Agent Creation Path**
   - Keep only `Team.SpawnAgent()` as the primary method
   - Make it handle both role-based and ad-hoc agent creation
   - Remove redundant `AddAgent()` and auto-spawning in `Call()`

2. **Unified Role Resolution**
   - Single method to resolve role configs with fallbacks
   - Consistent model selection logic
   - Proper prompt injection

### Phase 3: Consolidate Context Injection (Later)
1. **Single Context Builder**
   - Replace `buildContextMinimal()` with provider-based system
   - Token budgeting as described in PRODUCT.md
   - Agent-specific context profiles

### Phase 4: Consolidate Tool Registry (Later)  
1. **Unified Tool System**
   - Single registry initialization
   - Consistent builtin registration
   - Remove ad-hoc tool filtering

## Files That Need Changes

### Immediate (Phase 1):
- `internal/team/team.go` - Add missing `ListRoles() []string` method
- `internal/tool/builtins_team.go` - Add `available_roles` tool, update `team_status`

### Next (Phase 2):
- `internal/team/agents.go` - Consolidate agent creation
- `internal/team/team.go` - Remove redundant creation methods
- `cmd/agentry/agent.go` - Simplify buildAgent

### Later (Phases 3-4):
- Context injection files
- Tool registry initialization

## Recommended Approach

**Start with Phase 1 only** - this is the minimal fix to solve the immediate Agent0 role awareness issue without breaking existing functionality. The other phases can be tackled incrementally after validating that Phase 1 works correctly.

The key insight is that the functionality already exists (`Team.roles` has the data) - it's just not properly exposed through the `teamAPI` interface that the builtin tools use.
