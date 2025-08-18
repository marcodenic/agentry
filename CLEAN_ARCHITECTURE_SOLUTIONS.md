# Clean Architecture Solutions for Import Cycles

## Current Problem: Import Cycle
```
internal/tool ←→ internal/team
```

The current `teamAPI` interface approach is a **valid pattern** but implemented poorly. Here are the proper ways to solve this:

## Option 1: Dependency Inversion (Recommended)

### Step 1: Create Shared Interface Package
```
internal/
├── contracts/          # ← New shared package
│   └── team.go        # Interface definitions
├── team/              # Implementation
└── tool/              # Consumers
```

### Step 2: Move Interface to Neutral Territory
```go
// internal/contracts/team.go
package contracts

type TeamService interface {
    // Agent management
    SpawnedAgentNames() []string    // Currently running agents
    AvailableRoleNames() []string   // Roles from config
    
    // Communication
    SendMessage(ctx context.Context, from, to, msg string) error
    
    // Delegation  
    DelegateTask(ctx context.Context, role, task string) (string, error)
}
```

### Step 3: Clean Implementation
```go
// internal/team/team.go
func (t *Team) SpawnedAgentNames() []string { return t.names }
func (t *Team) AvailableRoleNames() []string { return t.ListRoleNames() }
func (t *Team) DelegateTask(ctx context.Context, role, task string) (string, error) {
    return t.Call(ctx, role, task)
}
```

### Step 4: Clean Consumer
```go
// internal/tool/builtins_team.go
import "github.com/marcodenic/agentry/internal/contracts"

func teamStatusTool() Tool {
    return func(ctx context.Context, args map[string]any) (string, error) {
        team := ctx.Value(TeamKey).(contracts.TeamService)
        spawned := team.SpawnedAgentNames()
        available := team.AvailableRoleNames()
        // Build response showing distinction
    }
}
```

## Option 2: Package Restructuring (Alternative)

Reorganize to eliminate the cycle entirely:

```
internal/
├── core/              # Agent, basic types
├── coordination/      # Team management (was team/)
├── tooling/          # Tool registry + builtins
└── runtime/          # Orchestrates everything
```

## Option 3: Event-Driven (Functional)

Replace interface with callback functions:

```go
// internal/tool/registry.go
type TeamCallbacks struct {
    GetSpawnedAgents func() []string
    GetAvailableRoles func() []string  
    DelegateTask func(ctx context.Context, role, task string) (string, error)
}

// Register callbacks when team is created
func RegisterTeamCallbacks(callbacks TeamCallbacks) {
    // Store in global registry or context
}
```

## Why Current Approach is Wrong

1. **Wrong Location**: Interface in consumer package (tool) instead of neutral ground
2. **Leaky Abstraction**: Exposes too many internals (GetSharedData, MarkMessagesAsRead)
3. **Poor Naming**: `Names()` vs `ListRoles()` - unclear distinction
4. **Type Mismatch**: Interface expects `[]string` but impl returns `[]*RoleConfig`
5. **No Enforcement**: Nothing ensures Team actually implements teamAPI

## My Recommendation: Option 1 (Dependency Inversion)

### Why This is Clean:
- ✅ **Single Responsibility**: Each package has one job
- ✅ **Clear Contract**: Interface defines exactly what tools need
- ✅ **Proper Naming**: Method names clarify spawned vs available
- ✅ **Type Safety**: Compiler enforces interface compliance
- ✅ **Testable**: Easy to mock `contracts.TeamService`

### Implementation Steps:
1. Create `internal/contracts/team.go` with clean interface
2. Update `internal/team/team.go` to implement interface
3. Update `internal/tool/builtins_team.go` to use interface
4. Update context storage to use interface type

This follows Go's **"Accept interfaces, return structs"** principle properly.

Want me to implement this clean architecture?
