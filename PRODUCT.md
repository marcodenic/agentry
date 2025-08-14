# Agentry Product Notes

## Command Line Usage

### Direct Prompt Invocation
You can invoke agentry with a direct prompt:
```bash
./agentry "your prompt here"
```

This runs the prompt through Agent 0 and returns the result. Useful for:
- Testing delegation scenarios
- Quick one-off tasks
- CI/CD integration
- Debugging agent behavior

Example:
```bash
./agentry "spawn a coder to review PRODUCT.md and report back"
```

### Available Commands
```bash
# Start TUI (default when no command provided)
./agentry
./agentry tui

# Direct prompt execution  
./agentry "create a hello world program"

# Command utilities
./agentry cost                          # Analyze cost from trace logs
./agentry analyze <trace-file>          # Analyze trace files
./agentry refresh-models                # Download latest model pricing
./agentry version                       # Show version
./agentry help                          # Show help

# Deprecated (will show warning)
./agentry chat                          # Use ./agentry instead
./agentry dev                           # Use ./agentry with AGENTRY_DEBUG=1
```

### Common Flags (TUI Mode)
```bash
./agentry --config path/to/.agentry.yaml
./agentry --theme dark
./agentry --save-id session1
./agentry --resume-id session1
```

### Debug Mode
Enable debug output with environment variable:
```bash
# Enable debug output
AGENTRY_DEBUG=1 ./agentry "test prompt"

# Debug with TUI (output goes to agentry.log)
AGENTRY_DEBUG=1 ./agentry
```

**Note**: In TUI mode, debug output is automatically redirected to avoid interfering with the interface.

## New Feature: Agent TODO List Tool

### Motivation
Agents need a way to keep track of tasks, decisions, and open threads during multi-step orchestration. A shared or per-agent TODO list tool will:
- Help agents plan, coordinate, and remember what needs to be done
- Allow agents to add, update, delete, and refer to TODO items
- Enable persistent memory of open tasks across agent runs
- Reduce cognitive load by externalizing task memory

### Proposed Tool API
```json
{
  "todo_add": {
    "description": "Add a new TODO item",
    "params": {
      "title": "string (required)",
      "description": "string (optional)",
      "priority": "low|medium|high (optional, default: medium)",
      "tags": "string[] (optional)",
      "agent_id": "string (optional, defaults to current agent)"
    }
  },
  "todo_list": {
    "description": "List TODO items with filtering",
    "params": {
      "agent_id": "string (optional, filter by agent)",
      "status": "pending|done|all (optional, default: pending)",
      "tags": "string[] (optional, filter by tags)",
      "limit": "number (optional, default: 10)"
    }
  },
  "todo_update": {
    "description": "Update a TODO item",
    "params": {
      "id": "string (required)",
      "status": "pending|done (optional)",
      "title": "string (optional)",
      "description": "string (optional)",
      "priority": "low|medium|high (optional)"
    }
  },
  "todo_delete": {
    "description": "Remove a TODO item",
    "params": {
      "id": "string (required)"
    }
  },
  "todo_get": {
    "description": "Get details for a specific TODO item",
    "params": {
      "id": "string (required)"
    }
  }
}
```

### Storage Design
- Store in `memstore` with namespace `todo:project:<project_path>`
- Each TODO has: id, title, description, priority, tags, agent_id, status, created_at, updated_at
- TODOs persist across agent runs and can be shared between agents
- Auto-cleanup of old completed TODOs (configurable TTL)

### Example Usage
```javascript
// Coder agent planning work
todo_add({
  title: "Refactor health endpoint",
  description: "Extract health check logic for better testability",
  priority: "high",
  tags: ["refactor", "health"]
})

// Planner reviewing status
todo_list({status: "pending", limit: 5})

// Agent completing work  
todo_update({id: "todo_123", status: "done"})
```

---

## Context Window Management & Token Budgeting

### Critical Problem Identified
Recent agent runs hit Anthropic's 30k token/minute rate limit due to **excessive context injection**. Analysis of current context injection:

**Token Breakdown (BEFORE agent's role prompt + task):**
- Hardcoded project details: ~200 tokens
- Workspace activity (3 events): ~500-1000 tokens  
- Team coordination history (3 events): ~500-1000 tokens
- Intelligence guidelines: ~400 tokens
- Role-specific instructions: ~800 tokens
- **TOTAL: 2,400-3,400 tokens of overhead**

### Problems with Current Approach
1. ✅ **Hardcoded assumptions**: ~~"You are in a Go project called 'agentry'" - not dynamic~~ **FIXED**: Now uses dynamic project detection via `buildRootFileTree()`
2. **No token budgeting**: No awareness of model's context window limits
3. **No relevance scoring**: All context treated equally important
4. **Additive only**: No way to truncate or prioritize context
5. **Rate limit breach**: Can exceed API limits before real work begins

### Solution Architecture
Implement intelligent context management:

```
Provider → Budget → Assembler → Agent
   ↓         ↓         ↓
Context    Token    Ranked
Sources    Limits   Context
```

**Context Budget by Model:**
- GPT-4o-mini: ~75% of 128k = 96k tokens (aggressive budget)
- Claude-3.5-Sonnet: ~75% of 200k = 150k tokens  
- Fallback: 75% of 4k = 3k tokens for older models

**Context Pack Providers:**
- Project structure (dynamic discovery)
- Recent activity (last N events, scored by relevance)
- Tool usage patterns
- Conversation history (summarized if needed)
- Error context (if recent failures)

**Relevance Scoring:**
- Task keywords match workspace files: +10
- Recent agent activity on similar tasks: +5  
- Coordination events involving current agent: +3
- Generic workspace info: +1

### Implementation Priority
1. **Remove hardcoded context** from `buildContextualInput()` 
2. **Add token counting** to context injection
3. **Implement context packs** with scoring
4. **Add budget enforcement** per model type
5. **Test with realistic workloads** across model providers

---

## Context Management Architecture

### Core Design: Provider → Budget → Assembler Pipeline

Replace hardcoded context injection with intelligent, token-aware context assembly:

**Context Packs** - Discrete, scored chunks of information:
- **TaskSpecProvider**: User request + agent role specifics
- **RulesProvider**: Project conventions, AGENT.md files 
- **WorkspaceSummaryProvider**: Dynamic project structure detection
- **ActiveFileProvider**: Current file with prefix/suffix windowing
- **RelatedFilesProvider**: Hybrid search (lexical + semantic + structural)
- **LSPDefsProvider**: Symbol definitions, references, hover docs
- **GitDiffProvider**: Staged/unstaged changes, commit context
- **TestFailProvider**: Recent test failures and error traces
- **RunOutputProvider**: Command outputs, build results
- **HistoryProvider**: Conversation history (compacted)
- **MemoryProvider**: Persistent project knowledge

**Agent Profiles** - Define which context packs each agent type gets:
```go
Profiles = {
    "coder": ["TaskSpec", "ActiveFile", "LSPDefs", "RelatedFiles", "GitDiff", "TestFail"],
    "planner": ["TaskSpec", "Rules", "WorkspaceSummary", "History", "Memory"],  
    "reviewer": ["GitDiff", "RelatedFiles", "Rules", "TestFail", "RunOutput"]
}
```

**Token Budgeting** - Enforce context window limits per model:
- Calculate available space: `modelCtx - system - userAsk - guardrails`
- Allocate budget by provider weights and task relevance
- Apply truncation strategies (prefix/suffix, outlines, excerpts)
- Always include provenance metadata (file:line references)

### Context Window Limits (from models_pricing.json)
- **Claude Models**: 200k tokens (aggressive budget: ~150k)
- **GPT Models**: 128k tokens (aggressive budget: ~96k) 
- **Fallback**: 8k tokens (conservative: ~6k)

Use `GetContextLimit(modelName)` from pricing table for accurate limits.

### File Selection Algorithm
Hybrid scoring for RelatedFiles provider:
```
score = 0.45 * semanticSim(task, fileEmb)
      + 0.25 * lexicalHits(tfidf/ripgrep density)  
      + 0.15 * structuralAffinity(import graph distance)
      + 0.10 * recency(recently edited/open)
      + 0.05 * centrality(call graph degree)
```

### LSP Integration Strategy
- Auto-start language servers based on detected project languages
- Cache definitions per file+version to avoid re-querying
- Include definition snippets (20-80 lines) + hover docs + reference sites
- Provide symbol context for code near cursor position

### Implementation Steps
1. **Replace buildContextualInput()** with Provider→Budget→Assembler pipeline
2. **Add ContextRegistry** to Team for pluggable providers
3. **Implement token counting** and budget enforcement per model
4. **Add agent profile mapping** (coder→coder profile, etc.)
5. **Integrate with existing memstore** for persistent memory

---

## Next Steps
- Implement the TODO tool for all agents
- **CRITICAL**: Refactor context injection to prevent rate limit violations
- Replace hardcoded context with dynamic, token-budgeted providers
- Test context window management across model providers
- Add LSP integration for richer code context
