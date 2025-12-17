# Agentry Reboot: Simplifying the Agent Architecture

_December 2025_

## Executive Summary

After honest assessment, the current multi-agent orchestration model in Agentry is overly complex and doesn't work well in practice. The emerging industry standard (Claude Code, Crush/OpenCode) uses a simpler model that's proving more effective. This document outlines the pivot.

---

## Current Architecture (What We Have)

### Hub-Spoke Multi-Agent Model

```
User Request
     ↓
┌─────────────┐
│  Agent 0    │ ← Coordinator/Orchestrator
│ (parent)    │ 
└──────┬──────┘
       │ delegation via "agent" tool
       ├──────────────────┬──────────────┐
       ↓                  ↓              ↓
┌──────────┐      ┌──────────┐   ┌──────────┐
│  coder   │      │ reviewer │   │  tester  │
│  agent   │      │  agent   │   │  agent   │
└──────────┘      └──────────┘   └──────────┘
```

### Problems

1. **Cognitive Overhead**: Agent 0 must decide which agent to delegate to, what instructions to give, and how to synthesize responses. This meta-planning layer is hard for LLMs.

2. **Sequential Bottleneck**: `Call()` blocks until the sub-agent completes. Only one delegation at a time.

3. **Over-specification**: Role-based agents (coder, reviewer, tester) force task categorization. Real work is messy and interleaved.

4. **Stateful Complexity**: Shared memory, coordination events, workspace events add latency and failure modes without clear benefit.

5. **15-Minute Timeouts**: Long-running delegations with complex timeout/recovery logic.

---

## Target Architecture (What Works)

### Single Agent + Parallel Read-Only Sub-Agents

```
User Request
     ↓
┌─────────────────────────────────────┐
│           Main Agent                │
│  (does ALL the actual work)         │
│  - reads files                      │
│  - writes files                     │
│  - runs commands                    │
│  - plans and iterates               │
└─────────────┬───────────────────────┘
              │
              │ "agent" tool (for parallel search)
              │
    ┌─────────┼─────────┐
    ↓         ↓         ↓
┌───────┐ ┌───────┐ ┌───────┐
│search │ │search │ │search │  ← Ephemeral, stateless
│agent 1│ │agent 2│ │agent 3│  ← Read-only tools only
└───────┘ └───────┘ └───────┘  ← Return in seconds
```

### Why This Works

| Aspect | Old Model | New Model |
|--------|-----------|-----------|
| Agent decision | "Who should do this?" | "Do I need parallel search?" |
| Sub-agent scope | Full capabilities | Read-only search |
| Sub-agent lifetime | Long-lived, stateful | Ephemeral, stateless |
| Parallelism | None (sequential) | Multiple concurrent |
| Coordination | Shared memory, events | Context window |
| Timeout | 15 minutes | Seconds |

### The Fundamental Insight

> The winning model isn't "multiple specialized agents collaborating" - it's **one smart agent with the ability to parallelize read operations**.

Why?
- LLMs are bad at multi-agent coordination
- LLMs are good at tool use in a tight loop  
- Context windows are now huge (200K+)
- The "orchestration" is implicit in the conversation history

---

## What To Keep

- **TUI** - Differentiator, good UX
- **Tracing/debugging infrastructure** - Valuable for development
- **Cost management** - Users need this
- **Tool implementations** - Core functionality
- **Streaming/conversation loop** - Works well

## What To Remove

- `internal/team/` - Complex team coordination
- `internal/teamruntime/` - Delegation machinery  
- Role-based agents (coder, reviewer, tester)
- Shared memory between agents
- Workspace events system
- 15-minute timeout logic
- Agent spawning/reuse

## What To Add

### Simple Agent Tool

```go
// internal/tool/subagent.go

// SubAgent spawns an ephemeral read-only agent for parallel search.
// It cannot modify files, run bash, or spawn further agents.
type SubAgent struct {
    client model.Client
}

// Available tools for sub-agents (read-only):
// - glob: Find files by pattern
// - grep: Search file contents  
// - view: Read file contents
// - ls: List directory
// - fetch: Get web content

// The sub-agent:
// 1. Receives a search prompt
// 2. Uses read-only tools to find information
// 3. Returns a text summary
// 4. Is discarded (no state retained)
```

### Tool Description (for LLM)

```markdown
Launch a new agent to perform search/research tasks. Use this when you need to 
find information across the codebase and want to parallelize the search.

**Capabilities:**
- Search files with glob patterns
- Search file contents with grep
- Read file contents
- List directories
- Fetch web content

**Limitations:**
- Cannot modify files
- Cannot run shell commands
- Cannot spawn further agents
- Stateless - each invocation starts fresh

**Usage:**
- Launch multiple agents concurrently for parallel search
- Be specific about what information to find
- The agent returns a text summary of findings

**When to use:**
- "Find all usages of function X" → spawn agent
- "Which files handle authentication?" → spawn agent
- "Search for config patterns" → spawn agent

**When NOT to use:**
- Reading a specific known file → use view directly
- Modifying files → do it yourself
- Running commands → do it yourself
```

---

## Implementation Plan

### Phase 1: Branch & Preserve
- [x] Commit current work (Google Gemini support, e2e test)
- [x] Create `reboot-simple-agent` branch

### Phase 2: Simplify Core
- [ ] Remove team/teamruntime packages (archive, don't delete repo history)
- [ ] Simplify Agent struct - remove role, delegation, shared memory
- [ ] Remove "agent" tool from main registry temporarily

### Phase 3: Add Simple Sub-Agent
- [ ] Create `internal/tool/subagent.go`
- [ ] Implement ephemeral agent with read-only tools
- [ ] Support parallel execution (multiple sub-agents at once)
- [ ] Add tool description optimized for LLM understanding

### Phase 4: Update Prompts
- [ ] Single system prompt (no role-based prompts)
- [ ] Clear tool descriptions
- [ ] Remove team/delegation instructions

### Phase 5: Test & Validate
- [ ] Run e2e test with new architecture
- [ ] Manual testing of common workflows
- [ ] Verify sub-agent parallelism works

---

## Success Criteria

1. **Simpler codebase**: Remove ~2000+ lines of team/delegation code
2. **Works for real tasks**: Can actually build things end-to-end
3. **Parallel search**: Multiple sub-agents run concurrently
4. **Fast sub-agents**: Return in seconds, not minutes
5. **No coordination overhead**: Main agent just uses tools

---

## Reference: How Crush Does It

From Crush's `agent_tool.md`:

```markdown
Launch a new agent that has access to: GlobTool, GrepTool, LS, View.

<usage_notes>
1. Launch multiple agents concurrently whenever possible
2. The agent returns a single message back to you
3. Each agent invocation is stateless
4. IMPORTANT: The agent cannot use Bash, Replace, Edit
</usage_notes>
```

From Crush's implementation:
- Uses `fantasy.NewParallelAgentTool` for concurrent execution
- Sub-agent gets `taskPrompt` (simpler than main coder prompt)
- Read-only tools: glob, grep, ls, view, sourcegraph
- Creates ephemeral session per invocation
- Returns text result directly

---

## Notes

This isn't abandoning the vision of AI-assisted development. It's recognizing that:

1. The best orchestration is a smart agent with good tools
2. Parallelism should be at the search/read level, not the "thinking" level
3. Simplicity beats sophistication when the simple thing works

The goal remains: a dependable assistant that can plan and execute development tasks. We're just using a better architecture to get there.
