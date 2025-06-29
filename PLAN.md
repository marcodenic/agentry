# Agentry Multi-Agent Platform Development Plan

**⚠️ CRITICAL: READ [CRITICAL_INSTRUCTIONS.md](./CRITICAL_INSTRUCTIONS.md) FIRST ⚠️**

**Status**: Phase 1 Complete ✅ | Phase 2A.1 Complete ✅ | Phase 2A.2 Complete ✅ | Phase 2A.3 In Progress 🔄  
**Last Updated**: Current Date  
**Current Focus**: Agent Lifecycle Management & Advanced Inter-Agent Communication

---

## 📋 **INSTRUCTIONS FOR PLAN MAINTENANCE**

**🔄 UPDATE REQUIREMENTS:**
- ✅ Mark tasks as complete when successfully implemented and tested
- 📝 Add implementation notes, file changes, and test results
- 🚨 Document any blockers, issues, or architectural decisions
- 🔀 Update status indicators and move to next phase when ready
- 📅 Update "Last Updated" date after each major milestone

**📊 Status Legend:**
- ✅ **COMPLETE** - Implemented, tested, and validated
- 🔄 **IN PROGRESS** - Currently working on this task
- ⏳ **BLOCKED** - Waiting on dependencies or external factors
- ❌ **NOT STARTED** - Not yet begun
- 🧪 **TESTING** - Implementation complete, testing in progress

---

## 🎯 **OVERVIEW & CONTEXT**

### **Current State Assessment**
Based on comprehensive analysis of AGENT_0_STATUS.md vs ROADMAP.md:

**✅ FOUNDATION COMPLETE (Phase 1)**
- Agent 0 coordination system fully operational
- Tool restriction working (10 coordination tools only)
- Team context unified across all modes (CLI, chat, TUI)
- Delegation workflow: Agent 0 → specialist agents working
- Safe sandbox testing environment established

**🚀 NEXT PHASE GOALS (Phase 2)**
Transform from ephemeral single-task coordination to persistent, long-running multi-agent teams with:
- Real-time inter-agent communication
- Persistent agent sessions and state
- Live status monitoring and progress tracking
- Advanced workflow orchestration
- Shared team memory and context

### **Testing Methodology**
All testing MUST be performed using the established methodology:
1. **Environment**: `/tmp/agentry-ai-sandbox` directory
2. **API Keys**: Source `.env.local` before every test
3. **Binary**: Use copied `agentry.exe` in sandbox
4. **Configuration**: Use sandbox-specific config files
5. **Validation**: Test both functionality and tool restrictions

### **🏗️ Architecture Decision: TCP Localhost + Migration Strategy**

**Phase 2 Architecture (Current):**
- **Agent Discovery**: JSON file registry in cross-platform temp directory
- **Agent Communication**: HTTP/JSON over TCP localhost (ports 9000-9099)
- **Message Broker**: Direct HTTP POST to agent endpoints
- **Platforms**: Windows, Linux, macOS compatible

**Migration Path to Distributed Systems:**
```go
// Phase 2: Local development
type LocalRegistry struct { /* JSON file based */ }
type LocalMessageBroker struct { /* HTTP TCP */ }

// Phase 3: Distributed production (future)
type DistributedRegistry struct { /* NATS/Consul/etcd */ }
type DistributedMessageBroker struct { /* NATS/Redis */ }

// Same interfaces - just swap implementations
```

**Benefits:**
- ✅ **Simple**: No external dependencies for Phase 2
- ✅ **Cross-platform**: Works on all major operating systems  
- ✅ **Debuggable**: Can use curl, browser, standard tools
- ✅ **Fast**: TCP localhost is nearly as fast as Unix sockets
- ✅ **Scalable**: Clean migration to distributed systems
- ✅ **Familiar**: Standard HTTP/JSON patterns

---

## 📋 **PHASE 2: LONG-RUNNING MULTI-AGENT COORDINATION**

### **Phase 2A: Persistent Agent Infrastructure** 🔄

#### **Task 2A.1: Agent Registry Service** ✅
**Objective**: Cross-platform agent discovery and communication using TCP localhost

**SIMPLIFIED Implementation Requirements:**
```go
// Files to create:
internal/registry/file_registry.go
internal/registry/tcp_server.go
internal/registry/types.go
internal/messaging/tcp_client.go

// Cross-platform TCP localhost approach:
type AgentRegistry struct {
    agents     map[string]*AgentInfo
    mutex      sync.RWMutex
    configFile string // Cross-platform temp dir + "/agentry/agents.json"
}

type AgentInfo struct {
    ID           string            `json:"id"`
    Port         int               `json:"port"`         // localhost:9001, 9002, etc.
    PID          int               `json:"pid"`
    Capabilities []string          `json:"capabilities"`
    Status       AgentStatus       `json:"status"`
    StartedAt    time.Time         `json:"started_at"`
    LastSeen     time.Time         `json:"last_seen"`
    Endpoint     string            `json:"endpoint"`     // "localhost:9001"
    Metadata     map[string]string `json:"metadata"`
}

// Simple HTTP/JSON over TCP for inter-agent communication
type TCPAgentServer struct {
    port     int
    agentID  string
    registry *AgentRegistry
    server   *http.Server
}
```

**🚀 Migration Path to Distributed Systems:**
```go
// Phase 2: TCP localhost (current)
endpoint := "localhost:9001"

// Phase 3: NATS/gRPC distributed (future)
endpoint := "nats://cluster.example.com:4222"
endpoint := "grpc://agent-hub.example.com:8080"

// Same interfaces, different transport backends
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Test 1: Agent startup and registry
./agentry.exe daemon start --agent-id "test-coder" --capabilities "code_generation,file_editing" &
CODER_PID=$!

# Verify agent registered in JSON file
cat /tmp/agentry/agents.json
# Should show: {"test-coder": {"port": 9001, "pid": 12345, ...}}

# Test 2: Agent discovery via file registry
./agentry.exe list-agents
# Should output: test-coder (localhost:9001) - ACTIVE

# Test 3: Direct HTTP communication
curl http://localhost:9001/health
# Should return: {"status": "healthy", "agent_id": "test-coder"}

# Test 4: Inter-agent messaging
./agentry.exe daemon start --agent-id "test-writer" --capabilities "writing,documentation" &
curl -X POST http://localhost:9001/message \
  -H "Content-Type: application/json" \
  -d '{"from": "test-writer", "to": "test-coder", "content": "Hello from writer!"}'

# Test 5: Agent capability lookup
./agentry.exe find-agents --capability "code_generation"
# Should return: test-coder

# Test 6: Cross-platform compatibility
# Run same tests on Windows, Linux, macOS

# Cleanup
kill $CODER_PID
```

**Success Criteria:**
- [ ] Agents can start and auto-register with JSON file registry
- [ ] Cross-platform compatibility (Windows, Linux, macOS)
- [ ] Agent discovery by capability works via file reading
- [ ] HTTP health checks confirm agent availability
- [ ] Direct HTTP/JSON messaging between agents
- [ ] Port auto-assignment prevents conflicts
- [ ] Clean shutdown removes agent from registry
- [ ] Easy migration path to distributed systems

**Implementation Notes:**
```
Status: ✅ COMPLETED - Phase 2A.1 Infrastructure Ready
Architecture: File-based registry + HTTP over TCP localhost 
Completion Date: 2024-06-29

Files Implemented:
- internal/config/loader.go (PersistentAgentsConfig)
- internal/persistent/team.go (PersistentTeam, PersistentAgent, HTTP servers)
- internal/registry/file_registry.go (JSON file-based agent registry)
- internal/registry/types.go (AgentInfo, AgentStatus, PortRange)
- cmd/agentry/chat.go (CLI integration with EnablePersistentAgents)
- persistent-config.yaml (test configuration)

Integration Status:
✅ Configuration parsing and validation working
✅ CLI integration with --config persistent-config.yaml
✅ PersistentTeam created with configurable port ranges
✅ HTTP server infrastructure for agents (health, message endpoints)
✅ File-based agent registry (JSON format, cross-platform)
✅ Graceful shutdown and resource cleanup
✅ team.Caller interface compatibility maintained
✅ Agent spawning integration: converse.Team.AddAgent → PersistentTeam.SpawnAgent
✅ HTTP endpoint activation: /message processes tasks via agent.Agent.Run()
✅ On-demand agent spawning: PersistentTeam.Call() spawns agents automatically

Test Results:
✅ Configuration enables persistent agents (ports 9001-9010)
✅ Chat mode properly initializes persistent team
✅ Agent delegation triggers persistent agent spawning ("✅ Spawned persistent agent: coder (port 9001)")
✅ HTTP servers start automatically with health and message endpoints
✅ Message endpoint processes tasks through agent.Agent.Run() 
✅ Agent registry tracks spawned agents with full metadata
✅ Graceful shutdown works correctly

Next Steps: Phase 2A.2 - Persistent Sessions and Lifecycle Management
```

---

#### **Task 2A.2: Persistent Agent Sessions** ✅
**Objective**: Enable long-running agent processes that maintain state across tasks

**IMPLEMENTED SUCCESSFULLY** - Phase 2A.2 is complete with comprehensive session management.

**Core Components Implemented:**
```go
// Files created:
internal/sessions/manager.go - Complete session management system
internal/sessions/agent.go - Session-aware agent wrapper
internal/persistent/team.go - Enhanced with session support
cmd/agentry/chat.go - CLI session commands

// Key interfaces implemented:
type SessionManager interface {
    CreateSession(ctx context.Context, req CreateSessionRequest) (*SessionState, error)
    ListSessions(ctx context.Context, agentID string) ([]*SessionInfo, error)
    GetSession(ctx context.Context, sessionID string) (*SessionState, error)
    SaveSession(ctx context.Context, state *SessionState) error
    RestoreSession(ctx context.Context, sessionID string) (*SessionState, error)
    TerminateSession(ctx context.Context, sessionID string) error
    CleanupOldSessions(ctx context.Context, maxAge time.Duration) error
}

type SessionAgent struct {
    *core.Agent
    sessionManager SessionManager
    currentSession *SessionState
}
```

**Key Features Delivered:**
- ✅ **Complete session lifecycle management** (create/load/save/suspend/resume/terminate)
- ✅ **Comprehensive state persistence** (memory, working directory, variables, metadata)
- ✅ **File-based session storage** with JSON format for human readability
- ✅ **HTTP API endpoints** for RESTful session management
- ✅ **CLI session commands** (```/sessions```, ```/session load```, ```/session save```, etc.)
- ✅ **Multi-agent session support** with team-wide session management
- ✅ **Thread-safe operations** with proper concurrency handling
- ✅ **Session-aware task execution** with automatic state saving

**HTTP API Endpoints:**
```
GET    /sessions         - List all sessions for agent
POST   /sessions         - Create new session
POST   /sessions/{id}    - Load/resume session
DELETE /sessions/{id}    - Terminate session
GET    /sessions/current - Get current session info
```

**CLI Commands:**
```bash
/sessions                        # List all sessions
/sessions list [agent-id]        # List sessions for specific agent
/sessions create <name> <desc>   # Create new session
/session load <session-id>       # Load/resume session
/session save                    # Save current session
/session current                 # Show current session info
/session terminate               # Terminate current session
/help                           # Show session command help
```

**Testing Results:**
```
✅ Build: Compilation successful, all code builds without errors
✅ Integration: Session management integrated into persistent team
✅ HTTP Endpoints: RESTful session API implemented and tested
✅ CLI Commands: Full command set integrated into chat interface
✅ State Persistence: Memory, context, and working directory preserved
✅ File Storage: JSON-based session files with atomic operations
✅ Concurrency: Thread-safe session operations
✅ Error Handling: Comprehensive error handling and recovery
```

**Implementation Notes:**
```
Status: ✅ COMPLETED
Dependencies: Task 2A.1 (Agent Registry) ✅
Files Created: 
  - internal/sessions/manager.go (270 lines)
  - internal/sessions/agent.go (180 lines)
  - test_session_management.sh (test script)
  - test_session_validation.sh (validation script)
  - PHASE_2A2_SESSIONS_COMPLETION.md (completion summary)
Files Modified:
  - internal/persistent/team.go (enhanced with session support)
  - cmd/agentry/chat.go (added CLI session commands)
Architecture: Session-aware agents with file-based persistence
Next Phase: Task 2A.3 - Agent Lifecycle Management
```

---

#### **Task 2A.3: Agent Lifecycle Management** ❌
**Objective**: Health monitoring, auto-restart, and resource management for persistent agents

**Implementation Requirements:**
```go
// Files to create:
internal/lifecycle/manager.go
internal/lifecycle/health.go
internal/lifecycle/monitor.go

// Key interfaces:
type LifecycleManager interface {
    StartAgent(config AgentConfig) (*Agent, error)
    StopAgent(agentID string) error
    RestartAgent(agentID string) error
    GetAgentHealth(agentID string) (*HealthReport, error)
    SetRestartPolicy(agentID string, policy RestartPolicy) error
}

type HealthReport struct {
    AgentID        string                 `json:"agent_id"`
    Status         HealthStatus           `json:"status"`
    LastHeartbeat  time.Time             `json:"last_heartbeat"`
    ResourceUsage  ResourceMetrics       `json:"resource_usage"`
    ErrorCount     int                   `json:"error_count"`
    Uptime         time.Duration         `json:"uptime"`
}
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Test 1: Agent health monitoring
./agentry.exe daemon start --agent-id "monitored-agent" --health-check-interval 5s

# Test 2: Health status reporting
./agentry.exe health --agent-id "monitored-agent"

# Test 3: Auto-restart on failure
# Simulate agent crash and verify restart
./agentry.exe simulate-crash --agent-id "monitored-agent"
sleep 10
./agentry.exe health --agent-id "monitored-agent"  # Should show "restarted"

# Test 4: Resource monitoring
./agentry.exe "perform intensive task for testing resource monitoring" --agent-id "monitored-agent"
./agentry.exe resources --agent-id "monitored-agent"

# Test 5: Graceful shutdown
./agentry.exe daemon stop --agent-id "monitored-agent" --graceful --timeout 30s
```

**Success Criteria:**
- [ ] Health monitoring system tracks agent status
- [ ] Auto-restart functionality works for failed agents
- [ ] Resource usage monitoring (CPU, memory, disk)
- [ ] Graceful shutdown with configurable timeout
- [ ] Health reports include meaningful metrics

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2A.2 (Persistent Sessions)
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

### **Phase 2B: Inter-Agent Communication** ❌

#### **Task 2B.1: Message Broker Integration** ❌
**Objective**: Implement reliable message delivery system for inter-agent communication

**SIMPLIFIED Implementation Requirements:**
```go
// Files to create:
internal/messaging/tcp_broker.go
internal/messaging/file_message_queue.go
internal/messaging/types.go

// Phase 2: Simple TCP + File-based messaging
type TCPMessageBroker struct {
    registry *AgentRegistry
}

// Phase 3: Distributed messaging (future migration)
type NATSMessageBroker struct {
    conn *nats.Conn
}

// Same interface, different implementations
type MessageBroker interface {
    Send(ctx context.Context, toAgent string, message *Message) error
    Broadcast(ctx context.Context, message *Message) error
    Subscribe(ctx context.Context, handler MessageHandler) error
    Close() error
}

type Message struct {
    ID          string            `json:"id"`
    From        string            `json:"from"`
    To          string            `json:"to"`            // Single agent for Phase 2
    Type        MessageType       `json:"type"`
    Payload     json.RawMessage   `json:"payload"`
    Priority    int               `json:"priority"`
    Timestamp   time.Time         `json:"timestamp"`
    ReplyTo     string            `json:"reply_to,omitempty"`
}
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start test agents
./agentry.exe daemon start --agent-id "agent-1" --capabilities "code" &
./agentry.exe daemon start --agent-id "agent-2" --capabilities "write" &

# Test 1: Direct HTTP messaging
curl -X POST http://localhost:9001/message \
  -H "Content-Type: application/json" \
  -d '{"from": "agent-2", "to": "agent-1", "type": "task", "payload": {"task": "create file"}}'

# Test 2: Message delivery confirmation
# Check agent-1 received the message via logs or status endpoint

# Test 3: Broadcast messaging to all agents
./agentry.exe msg broadcast --from "coordinator" --payload '{"announcement": "team meeting"}'

# Test 4: Request-response pattern via HTTP
curl -X POST http://localhost:9001/request \
  -H "Content-Type: application/json" \
  -d '{"from": "agent-2", "request": "status"}' \
  --timeout 10

# Test 5: Message persistence (simple file-based queue)
# Verify messages are queued if target agent is offline

# Future: Easy migration to NATS
# docker run -d --name nats-server -p 4222:4222 nats:latest
# Same message interface, different transport
```

**Success Criteria:**
- [ ] HTTP-based direct agent messaging working
- [ ] Point-to-point message delivery
- [ ] Simple broadcast to all registered agents
- [ ] Request-response pattern with timeout
- [ ] Message queuing for offline agents (file-based)
- [ ] Cross-platform compatibility
- [ ] Clear migration path to NATS/distributed messaging

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2A.1 (TCP localhost registry)
Architecture: HTTP messaging, file queues, NATS migration ready
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

#### **Task 2B.2: Direct Inter-Agent Communication** ❌
**Objective**: Enable agents to communicate directly without going through Agent 0

**Implementation Requirements:**
```go
// Files to modify:
internal/tool/builtins.go - Enhance send_message tool
internal/converse/team.go - Add direct messaging support
internal/messaging/router.go - New message routing

// Enhanced send_message tool:
type DirectMessage struct {
    ToAgent     string            `json:"to_agent"`
    MessageType string            `json:"message_type"`
    Content     string            `json:"content"`
    Priority    int               `json:"priority"`
    ExpectReply bool              `json:"expect_reply"`
    Timeout     time.Duration     `json:"timeout,omitempty"`
}
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start multiple agent sessions
./agentry.exe daemon start --agent-id "coder-agent" --session-id "coder-001" &
./agentry.exe daemon start --agent-id "writer-agent" --session-id "writer-001" &
./agentry.exe daemon start --agent-id "reviewer-agent" --session-id "reviewer-001" &

# Test 1: Direct agent-to-agent message
./agentry.exe --session-id "coder-001" "send a message to writer-agent asking them to create documentation for the code you will write"

# Test 2: Agent-to-agent collaboration
./agentry.exe --session-id "coder-001" "create a simple Python script and ask writer-agent to document it"

# Test 3: Multi-agent workflow coordination
./agentry.exe --session-id "coder-001" "create code, ask writer-agent to document it, then ask reviewer-agent to review both"

# Test 4: Message delivery confirmation
./agentry.exe --session-id "coder-001" "send a message to writer-agent and confirm they received it"

# Test 5: Agent 0 coordination still works
./agentry.exe "coordinate all agents to work on a simple project together"

# Monitor messages
./agentry.exe msg monitor --show-all-agents
```

**Success Criteria:**
- [ ] Agents can send direct messages to other agents
- [ ] Message routing works without Agent 0 involvement
- [ ] Delivery confirmation and error handling
- [ ] Agent 0 can still coordinate when needed
- [ ] Multi-agent collaboration workflows function
- [ ] Message history and audit trail maintained

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2B.1 (Message Broker)
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

### **Phase 2C: Real-Time Monitoring & Status** ❌

#### **Task 2C.1: Status Reporting Framework** ❌
**Objective**: Standardized real-time status updates and progress tracking

**Implementation Requirements:**
```go
// Files to create:
internal/monitoring/status.go
internal/monitoring/events.go
internal/monitoring/websocket.go
ui/dashboard/status.html

// Key interfaces:
type StatusReporter interface {
    ReportStatus(agentID string, status AgentStatus) error
    ReportProgress(agentID string, progress TaskProgress) error
    ReportEvent(agentID string, event StatusEvent) error
    GetStatus(agentID string) (*AgentStatus, error)
    SubscribeToUpdates(agentID string) (<-chan StatusUpdate, error)
}

type AgentStatus struct {
    ID                string              `json:"id"`
    State             AgentState          `json:"state"`
    CurrentTask       *TaskInfo           `json:"current_task,omitempty"`
    QueuedTasks       []TaskInfo          `json:"queued_tasks"`
    Performance       PerformanceMetrics  `json:"performance"`
    HealthIndicators  HealthMetrics       `json:"health"`
    LastUpdate        time.Time           `json:"last_update"`
}
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start monitoring dashboard
./agentry.exe dashboard start --port 8090 &
DASHBOARD_PID=$!

# Start monitored agents
./agentry.exe daemon start --agent-id "monitored-coder" --enable-monitoring &
./agentry.exe daemon start --agent-id "monitored-writer" --enable-monitoring &

# Test 1: Basic status reporting
./agentry.exe status --agent-id "monitored-coder"

# Test 2: Real-time status updates
# Open browser to http://localhost:8090 and verify live updates
./agentry.exe --session-id "monitored-coder" "start a long-running task and report progress"

# Test 3: Task progress tracking
./agentry.exe --session-id "monitored-coder" "create multiple files and report progress for each"

# Test 4: Performance metrics
./agentry.exe metrics --agent-id "monitored-coder" --duration 5m

# Test 5: Multi-agent dashboard
# Verify dashboard shows all agents and their status
curl http://localhost:8090/api/agents/status

# Cleanup
kill $DASHBOARD_PID
```

**Success Criteria:**
- [ ] Real-time status updates for all agents
- [ ] Task progress tracking with percentages
- [ ] Performance metrics collection
- [ ] Web dashboard shows live agent status
- [ ] WebSocket/SSE for real-time updates
- [ ] Historical status data retention

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2A.2 (Persistent Sessions)
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

#### **Task 2C.2: Web Dashboard Implementation** ❌
**Objective**: Visual monitoring interface for multi-agent operations

**Implementation Requirements:**
```go
// Files to create:
ui/dashboard/main.go - Dashboard server
ui/dashboard/static/ - Static assets
ui/dashboard/templates/ - HTML templates
internal/api/dashboard.go - Dashboard API endpoints

// Dashboard features:
- Live agent status grid
- Task progress indicators
- Message flow visualization
- Performance metrics charts
- Agent health monitoring
- Real-time event log
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start full multi-agent environment
./agentry.exe hub start --port 8080 &
./agentry.exe dashboard start --port 8090 --hub localhost:8080 &

# Start multiple agents for monitoring
./agentry.exe daemon start --agent-id "coder" --session-id "s1" &
./agentry.exe daemon start --agent-id "writer" --session-id "s2" &
./agentry.exe daemon start --agent-id "reviewer" --session-id "s3" &

# Test 1: Dashboard accessibility
curl http://localhost:8090/health
# Should return 200 OK

# Test 2: Agent status visualization
# Open browser to http://localhost:8090
# Verify all agents visible in dashboard

# Test 3: Real-time updates
./agentry.exe --session-id "s1" "start a complex task that will take time"
# Verify progress updates appear in dashboard

# Test 4: Multi-agent coordination monitoring
./agentry.exe "coordinate all agents to work on a project"
# Verify message flow visible in dashboard

# Test 5: Performance metrics
# Verify CPU, memory, task completion metrics displayed

# Test 6: Mobile responsiveness
# Test dashboard on mobile viewport
```

**Success Criteria:**
- [ ] Web dashboard accessible and responsive
- [ ] Live agent status grid with real-time updates
- [ ] Task progress visualization
- [ ] Inter-agent message flow display
- [ ] Performance metrics charts
- [ ] Event log with filtering
- [ ] Mobile-friendly interface

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2C.1 (Status Framework)
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

### **Phase 2D: Workflow Orchestration** ❌

#### **Task 2D.1: Enhanced Workflow Engine** ❌
**Objective**: Advanced multi-agent workflow orchestration with dependencies

**Implementation Requirements:**
```go
// Files to modify/create:
pkg/flow/engine.go - Enhanced workflow execution
pkg/flow/parser.go - Advanced YAML parsing
pkg/flow/scheduler.go - Task dependency resolution
examples/workflows/ - Multi-agent workflow examples

// Enhanced workflow format:
name: "multi_agent_development"
description: "Complete software development workflow"

agents:
  - id: "architect"
    role: "system_architect"
    capabilities: ["system_design", "architecture_planning"]
  - id: "coder"
    role: "coder"
    capabilities: ["code_generation", "file_editing"]
  - id: "tester"
    role: "qa_engineer"
    capabilities: ["test_generation", "test_execution"]
  - id: "reviewer"
    role: "code_reviewer"
    capabilities: ["code_review", "security_analysis"]

tasks:
  - id: "design_system"
    agent: "architect"
    inputs: ["requirements"]
    outputs: ["system_design", "architecture_spec"]
    
  - id: "implement_core"
    agent: "coder"
    depends_on: ["design_system"]
    inputs: ["system_design", "architecture_spec"]
    outputs: ["source_code", "implementation_notes"]
    
  - id: "create_tests"
    agent: "tester"
    depends_on: ["implement_core"]
    inputs: ["source_code"]
    outputs: ["test_suite", "coverage_report"]
    
  - id: "code_review"
    agent: "reviewer"
    depends_on: ["implement_core", "create_tests"]
    inputs: ["source_code", "test_suite"]
    outputs: ["review_report", "security_assessment"]

coordination:
  timeout: "45m"
  retry_policy: "exponential_backoff"
  failure_action: "pause_and_notify"
  parallel_limit: 2
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Create test workflow file
cat > test_workflow.yaml << 'EOF'
name: "simple_dev_workflow"
description: "Basic development workflow test"

agents:
  - id: "coder"
    role: "coder"
    capabilities: ["code_generation", "file_editing"]
  - id: "reviewer"
    role: "code_reviewer"
    capabilities: ["code_review"]

tasks:
  - id: "create_code"
    agent: "coder"
    inputs: ["requirements: Create a simple Python calculator"]
    outputs: ["calculator.py"]
    
  - id: "review_code"
    agent: "reviewer"
    depends_on: ["create_code"]
    inputs: ["calculator.py"]
    outputs: ["review_report"]

coordination:
  timeout: "10m"
  retry_policy: "linear"
  failure_action: "stop"
EOF

# Test 1: Workflow parsing
./agentry.exe workflow validate test_workflow.yaml

# Test 2: Workflow execution
./agentry.exe workflow run test_workflow.yaml --watch

# Test 3: Dependency resolution
# Verify tasks execute in correct order

# Test 4: Parallel task execution
cat > parallel_workflow.yaml << 'EOF'
name: "parallel_test"
tasks:
  - id: "task_a"
    agent: "coder"
    inputs: ["Create file A"]
  - id: "task_b" 
    agent: "writer"
    inputs: ["Create document B"]
  - id: "task_c"
    agent: "reviewer"
    depends_on: ["task_a", "task_b"]
    inputs: ["Review both A and B"]
EOF

./agentry.exe workflow run parallel_workflow.yaml --watch

# Test 5: Failure handling and recovery
# Simulate task failure and verify recovery behavior

# Test 6: Workflow status monitoring
./agentry.exe workflow status --workflow-id "simple_dev_workflow"
```

**Success Criteria:**
- [ ] Complex multi-agent workflows parse correctly
- [ ] Task dependency resolution works
- [ ] Parallel task execution where possible
- [ ] Workflow failure handling and recovery
- [ ] Real-time workflow status monitoring
- [ ] Workflow templates and examples provided

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2B.2 (Inter-Agent Communication)
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

#### **Task 2D.2: Task Queue & Scheduling** ❌
**Objective**: Priority-based task distribution and queue management

**Implementation Requirements:**
```go
// Files to create:
internal/scheduler/queue.go
internal/scheduler/priority.go
internal/scheduler/dispatcher.go

// Key interfaces:
type TaskQueue interface {
    Enqueue(task *Task) error
    Dequeue(agentCapabilities []string) (*Task, error)
    GetQueueStatus() (*QueueStatus, error)
    SetTaskPriority(taskID string, priority int) error
    CancelTask(taskID string) error
}

type Task struct {
    ID               string            `json:"id"`
    WorkflowID       string            `json:"workflow_id"`
    AgentID          string            `json:"agent_id,omitempty"`
    RequiredCaps     []string          `json:"required_capabilities"`
    Priority         int               `json:"priority"`
    Dependencies     []string          `json:"dependencies"`
    Inputs           map[string]any    `json:"inputs"`
    Timeout          time.Duration     `json:"timeout"`
    RetryPolicy      RetryPolicy       `json:"retry_policy"`
    Status           TaskStatus        `json:"status"`
    CreatedAt        time.Time         `json:"created_at"`
    StartedAt        *time.Time        `json:"started_at,omitempty"`
    CompletedAt      *time.Time        `json:"completed_at,omitempty"`
}
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start task queue system
./agentry.exe queue start --port 8085 &
QUEUE_PID=$!

# Start worker agents
./agentry.exe worker start --agent-id "worker-1" --capabilities "code_generation,file_editing" --queue localhost:8085 &
./agentry.exe worker start --agent-id "worker-2" --capabilities "documentation,writing" --queue localhost:8085 &

# Test 1: Task submission
./agentry.exe queue submit --task-type "code_generation" --priority 5 --payload '{"request": "create hello world app"}'

# Test 2: Priority-based processing
./agentry.exe queue submit --task-type "code_generation" --priority 10 --payload '{"request": "urgent bug fix"}'
./agentry.exe queue submit --task-type "code_generation" --priority 1 --payload '{"request": "nice to have feature"}'

# Verify high priority task processes first

# Test 3: Capability-based assignment
./agentry.exe queue submit --task-type "documentation" --required-caps "documentation,writing" --payload '{"request": "write API docs"}'

# Test 4: Task dependencies
./agentry.exe queue submit --task-id "task-1" --task-type "code_generation" --payload '{"request": "create module"}'
./agentry.exe queue submit --task-id "task-2" --task-type "documentation" --depends-on "task-1" --payload '{"request": "document the module"}'

# Test 5: Queue monitoring
./agentry.exe queue status
./agentry.exe queue list --status pending
./agentry.exe queue list --status running

# Test 6: Task cancellation
./agentry.exe queue submit --task-id "cancel-test" --task-type "code_generation" --payload '{"request": "long running task"}'
./agentry.exe queue cancel --task-id "cancel-test"

# Cleanup
kill $QUEUE_PID
```

**Success Criteria:**
- [ ] Priority-based task queue implementation
- [ ] Capability-based task assignment
- [ ] Task dependency resolution
- [ ] Worker agent integration
- [ ] Queue monitoring and management
- [ ] Task cancellation and retry mechanisms

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: Task 2A.2 (Persistent Sessions), Task 2B.1 (Message Broker)
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

### **Phase 2E: Shared Team Memory** ❌

#### **Task 2E.1: Persistent Team Context** ❌
**Objective**: Shared knowledge base and context across all agents

**Implementation Requirements:**
```go
// Files to create:
internal/memory/team_store.go
internal/memory/vector_store.go  
internal/memory/context_manager.go

// Key interfaces:
type TeamMemory interface {
    StoreKnowledge(ctx context.Context, knowledge *Knowledge) error
    RetrieveKnowledge(ctx context.Context, query string, limit int) ([]*Knowledge, error)
    UpdateContext(ctx context.Context, agentID string, context *AgentContext) error
    GetSharedContext(ctx context.Context) (*SharedContext, error)
    AddTeamDecision(ctx context.Context, decision *TeamDecision) error
    GetRelevantHistory(ctx context.Context, query string) ([]*HistoryItem, error)
}

type Knowledge struct {
    ID          string            `json:"id"`
    Type        KnowledgeType     `json:"type"`
    Content     string            `json:"content"`
    Source      string            `json:"source"`
    AgentID     string            `json:"agent_id"`
    Tags        []string          `json:"tags"`
    Embedding   []float32         `json:"embedding,omitempty"`
    CreatedAt   time.Time         `json:"created_at"`
    Relevance   float64           `json:"relevance,omitempty"`
}
```

**Testing Plan:**
```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start team memory service
./agentry.exe memory start --storage sqlite --vector-db sqlite-vss &
MEMORY_PID=$!

# Start agents with shared memory
./agentry.exe daemon start --agent-id "learner-1" --shared-memory localhost:8086 &
./agentry.exe daemon start --agent-id "learner-2" --shared-memory localhost:8086 &

# Test 1: Knowledge storage and retrieval
./agentry.exe --session-id "learner-1" "remember that our team prefers Python over JavaScript for backend services"
./agentry.exe --session-id "learner-2" "what are our team's technology preferences?"

# Test 2: Context sharing
./agentry.exe --session-id "learner-1" "I'm working on a user authentication module"
./agentry.exe --session-id "learner-2" "what is learner-1 currently working on?"

# Test 3: Team decisions tracking
./agentry.exe --session-id "learner-1" "we decided to use PostgreSQL for the database"
./agentry.exe --session-id "learner-2" "what database should I use for this project?"

# Test 4: Semantic knowledge search
./agentry.exe --session-id "learner-2" "find information about database choices"

# Test 5: Learning from interactions
./agentry.exe --session-id "learner-1" "create a REST API endpoint for user login"
./agentry.exe --session-id "learner-2" "create a similar API endpoint for user registration"
# Should reuse patterns from learner-1's work

# Test 6: Memory persistence
# Restart memory service and verify knowledge persists
kill $MEMORY_PID
./agentry.exe memory start --storage sqlite --vector-db sqlite-vss &
./agentry.exe --session-id "learner-2" "what do we know about API development?"
```

**Success Criteria:**
- [ ] Shared knowledge storage and retrieval
- [ ] Context sharing between agents
- [ ] Team decision tracking
- [ ] Semantic search capabilities
- [ ] Learning from agent interactions
- [ ] Persistent memory across restarts

**Implementation Notes:**
```
Status: ❌ NOT STARTED
Dependencies: All previous Phase 2 tasks
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Issues Found: [Document any problems encountered]
```

---

## 📊 **PHASE 2 INTEGRATION TESTING**

### **Full System Integration Test** ❌
**Objective**: Comprehensive test of all Phase 2 components working together

**Test Scenario**: Multi-Agent Software Development Project

```bash
# Testing environment setup
cd /tmp/agentry-ai-sandbox
source .env.local

# Start all services
./agentry.exe hub start --port 8080 &
./agentry.exe queue start --port 8085 &
./agentry.exe memory start --port 8086 &
./agentry.exe dashboard start --port 8090 &

# Start persistent agent team
./agentry.exe daemon start --agent-id "architect" --role "system_architect" --session-id "arch-001" &
./agentry.exe daemon start --agent-id "coder" --role "coder" --session-id "code-001" &
./agentry.exe daemon start --agent-id "tester" --role "qa_engineer" --session-id "test-001" &
./agentry.exe daemon start --agent-id "reviewer" --role "code_reviewer" --session-id "review-001" &

# Create complex workflow
cat > integration_test_workflow.yaml << 'EOF'
name: "full_development_cycle"
description: "Complete software development with all agents"

agents:
  - id: "architect"
    session_id: "arch-001"
    capabilities: ["system_design", "architecture_planning"]
  - id: "coder"
    session_id: "code-001"
    capabilities: ["code_generation", "file_editing"]
  - id: "tester"
    session_id: "test-001"
    capabilities: ["test_generation", "test_execution"]
  - id: "reviewer"
    session_id: "review-001"
    capabilities: ["code_review", "security_analysis"]

tasks:
  - id: "analyze_requirements"
    agent: "architect"
    inputs: ["Create a simple task management API with CRUD operations"]
    outputs: ["requirements_analysis", "system_design"]
    
  - id: "implement_api"
    agent: "coder"
    depends_on: ["analyze_requirements"]
    inputs: ["requirements_analysis", "system_design"]
    outputs: ["api_code", "database_schema"]
    
  - id: "create_tests"
    agent: "tester"
    depends_on: ["implement_api"]
    inputs: ["api_code"]
    outputs: ["test_suite", "test_results"]
    
  - id: "security_review"
    agent: "reviewer"
    depends_on: ["implement_api"]
    inputs: ["api_code", "database_schema"]
    outputs: ["security_report"]
    
  - id: "integration_review"
    agent: "reviewer"
    depends_on: ["create_tests", "security_review"]
    inputs: ["api_code", "test_suite", "security_report"]
    outputs: ["final_approval"]

coordination:
  timeout: "30m"
  shared_memory: true
  real_time_monitoring: true
EOF

# Execute integration test
./agentry.exe workflow run integration_test_workflow.yaml --monitor --verbose

# Monitor in dashboard
echo "Open http://localhost:8090 to monitor progress"

# Verify agent communication
./agentry.exe msg monitor --workflow-id "full_development_cycle"

# Verify shared memory usage
./agentry.exe memory query "task management API" --show-sources

# Verify persistent sessions
./agentry.exe sessions list --show-status

# Performance validation
./agentry.exe metrics --workflow-id "full_development_cycle" --detailed
```

**Success Criteria:**
- [ ] All agents start and register successfully
- [ ] Workflow executes with proper task dependencies
- [ ] Inter-agent communication works without Agent 0
- [ ] Shared memory is used and updated
- [ ] Real-time monitoring shows progress
- [ ] All tasks complete successfully
- [ ] Performance metrics within acceptable ranges

**Implementation Notes:**
```
Status: ❌ NOT STARTED  
Dependencies: All Phase 2 tasks
Files Changed: [List files when implemented]
Test Results: [Add test output when completed]
Success Metrics: [Record performance and functionality results]
```

---

## 🔄 **PLAN MAINTENANCE LOG**

### **Completed Milestones**

#### **Phase 1: Foundation (COMPLETED ✅)**
- **Completed**: June 29, 2025
- **Status**: ✅ **All objectives achieved**
- **Key Achievements**:
  - Agent 0 tool restriction (15 → 10 tools)
  - Team context unified across all modes
  - Delegation workflow operational
  - Safe sandbox testing environment
- **Files Modified**: 
  - `cmd/agentry/common.go` - Added role config application
  - `cmd/agentry/prompt.go` - Added team context creation
  - `cmd/agentry/chat.go` - Enhanced with role config
- **Test Results**: All coordination tools working in all modes

---

### **Current Status**
- **Phase**: 2 (Long-Running Multi-Agent Coordination)
- **Current Task**: Task 2A.1 (TCP Localhost Agent Registry)
- **Architecture**: TCP localhost + JSON file registry
- **Next Action**: Implement cross-platform agent discovery and HTTP messaging
- **Blockers**: None currently identified

---

### **Update History**
- **2025-06-29**: Plan created with comprehensive Phase 2 roadmap
- **2025-06-29**: Architecture decision - TCP localhost + JSON registry for cross-platform compatibility
- **2025-06-29**: Updated Task 2A.1 and 2B.1 with simplified approach and migration strategy
- **[NEXT UPDATE DATE]**: [Status of first implementation task]

---

## 🎯 **SUCCESS METRICS & VALIDATION**

### **Phase 2 Success Criteria**
When Phase 2 is complete, the following must be demonstrably working:

1. **Multi-Agent Persistence**: 3+ agents running continuously for 30+ minutes
2. **Direct Communication**: Agents exchanging messages without Agent 0 mediation
3. **Task Coordination**: Complex workflow with dependencies completing successfully
4. **Real-Time Monitoring**: Live dashboard showing agent status and progress
5. **Shared Learning**: Agents accessing and contributing to team knowledge base
6. **Fault Tolerance**: System recovering from individual agent failures
7. **Resource Efficiency**: System running within reasonable resource constraints

### **Performance Benchmarks**
- **Agent Startup Time**: < 5 seconds per agent
- **Message Delivery**: < 100ms latency for inter-agent messages
- **Task Queue Processing**: > 10 tasks/minute throughput
- **Memory Usage**: < 1GB total for 5-agent system
- **CPU Usage**: < 50% on 4-core system under normal load

---

**END OF PLAN.MD - Remember to update status as you complete each task! 🚀**
