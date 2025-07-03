# Agent Registry Architecture Decision

## Context
We need to implement Phase 2A.1: Agent Registry Service for local agent discovery and lifecycle management.

## Decision: HTTP REST over gRPC for Phase 2

### Current Requirements Analysis:
- **Scope**: Local agents on same machine
- **Use Cases**: Agent registration, discovery, status tracking
- **Testing**: All scenarios are localhost-based
- **Timeline**: Need rapid prototyping and testing

### Trade-off Analysis:

#### gRPC Approach (Current Implementation)
**Pros:**
- Type safety with protobuf
- Future-ready for distributed scenarios
- Streaming capabilities
- Multi-language support

**Cons:**
- Complexity overhead for local-only communication
- Requires protoc toolchain
- Network overhead even locally
- More complex debugging and testing
- Over-engineering for current phase

#### HTTP REST Approach (Recommended)
**Pros:**
- Simple implementation and testing
- Easy CLI integration with standard HTTP
- Lightweight for local communication
- Better debugging (curl-friendly)
- Faster iteration and development
- Can migrate to gRPC later if needed

**Cons:**
- Less type safety (but manageable with good structs)
- Manual JSON serialization (but Go handles this well)
- Need to implement streaming separately if required later

## Recommended Implementation

### 1. HTTP Server Structure
```go
// internal/registry/http_server.go
type HTTPRegistryServer struct {
    registry *AgentRegistry
    server   *http.Server
}

// Routes:
// POST   /agents                    - Register agent
// DELETE /agents/{id}               - Deregister agent
// GET    /agents                    - List all agents
// GET    /agents/{id}               - Get agent details
// GET    /agents/find?capability=X  - Find agents by capability
// POST   /agents/{id}/heartbeat     - Agent heartbeat
// GET    /status                    - Registry health
```

### 2. JSON Messages
```go
type RegisterAgentRequest struct {
    ID           string            `json:"id"`
    Capabilities []string          `json:"capabilities"`
    Endpoint     string            `json:"endpoint"`
    Metadata     map[string]string `json:"metadata"`
}

type AgentResponse struct {
    ID           string            `json:"id"`
    Capabilities []string          `json:"capabilities"`
    Endpoint     string            `json:"endpoint"`
    Status       string            `json:"status"`
    LastSeen     time.Time         `json:"last_seen"`
    Metadata     map[string]string `json:"metadata"`
}
```

### 3. CLI Commands
```bash
# Much simpler HTTP-based CLI commands
./agentry register-agent --id "coder-1" --capabilities "code,file" --endpoint "localhost:8081"
./agentry list-agents
./agentry find-agents --capability "code"
./agentry agent-status --id "coder-1"
```

## Migration Path
If we later need gRPC for distributed scenarios:
1. Keep the same internal registry interfaces
2. Add gRPC server alongside HTTP server
3. Migrate clients gradually
4. The registry core logic remains unchanged

## Decision
**Implement HTTP REST for Phase 2A.1**, keeping the existing registry core logic but replacing the gRPC transport layer.

This allows us to:
- Move faster on Phase 2 implementation
- Have simpler testing and debugging
- Maintain flexibility for future gRPC migration
- Focus on the core registry functionality rather than transport complexity
