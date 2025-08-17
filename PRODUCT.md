# Agentry Product & Roadmap

Single authoritative doc consolidating former PLAN.md + FEATURES.md + forward architecture notes. Keep terse, actionable. (Update whenever shipping or re‑prioritizing.)

## Vision (Condensed)
Local‑first, observable, resilient multi‑agent orchestration: a small, composable Go core that can:
- Spawn and coordinate role agents safely (permissioned tools, sandbox pending).
- Persist working/semantic memory across runs.
- Provide transparent cost / token / trace observability.
- Be easy to extend (tools, model backends, context providers) without forking.

## Current Foundations (What Exists)
- Core loop: tool calling, tracing, cost+token accounting, error resilience (treat errors as results, retry caps).
- Tools: 30+ built‑ins (atomic file ops, search/replace, web/network, OpenAPI, MCP, audit, patch, delegation, etc.).
- Models: OpenAI, Anthropic, mock; unified `model.Client` interface.
- Multi‑agent: Team registry + delegation tool; role templates loaded from file system.
- Memory: per‑agent in‑proc convo history + vector store (shared across spawned agents) + SharedStore (memory/file) for persistence + checkpoints.
- TUI: streaming, delegation, token/cost bar, input history, logo gradient, diagnostics summary, safe autoscroll.
- CLI: JSON‑pure automation commands (`invoke`, `team`, `memory`, `analyze`, `refresh-models`). Deprecated: `chat`, `dev`.
- Platform context + available roles injected into system prompt (legacy platform injection still active).

## Recently Completed (Highlights)
- SharedStore (memory + file backends) & TTL/GC; coordination events persisted.
- Inbox built‑ins: `inbox_read`, `inbox_clear`, `request_help`, `workspace_events` + automatic unread injection per turn.
- Team built‑ins use real Team context; delegation safety (worker agents lose `agent` tool).
- LSP diagnostics parsing (gopls / tsc) surfaced in TUI summary.
- Minimal context builder (removed large hardcoded blocks) – awaiting v2 pipeline.
- JSON stdout purity for automation commands; human logs to stderr.
- Pricing cache path moved to user cache dir; model pricing refresh command.
- Iteration cap removal (agents run until final answer) with optional env budget stop.

## Hardening & Cleanup (Code Tightening – No New User Features)
Architecture
- [ ] Context v2: Provider → Budget → Assembler (replace implicit logic in `BuildMessages`).
- [x] Extract tool execution segment from `Agent.Run` → `executeToolCalls` (testable, smaller cyclomatic complexity).
- [x] Add cancellation checks in agent loop (before/after model + each tool).
- [ ] Introduce `AgentConfig` struct (consolidate scattered env usage: budgets, error handling, model name).
- [x] Provide fallback minimal system prompt if role file missing (avoid empty system message). (DONE)
- [ ] Replace UUID Prometheus token label with role (`agent_role`) to avoid high cardinality.
Code Quality
- [x] Consolidate env helpers (`getenvInt`, etc.) into `internal/env` package; remove duplicates.
- [x] Simplify `compactHistory` using `sort.Slice` (remove O(n^2) manual sort). (DONE)
- [ ] Clarify Spawn semantics (vector store sharing) with expanded comment + potential option to isolate.
- [ ] Collapse `defaultPrompt()` + `GetDefaultPrompt()` duplication; keep one public function.
- [ ] Guard verbose debug prompt/tool dumps behind selectively enabled debug channels.
- [ ] Normalize model name format (`provider/model`) with a helper.
Observability
- [x] Add model latency histogram; add tool error counter (low cardinality). (Histogram DONE; error counter TODO)
- [ ] Stream trace chunks in larger groups (reduce per‑char overhead) while preserving UI responsiveness.
Testing
- [ ] Unit: tool execution error recovery (consecutive error cap), history compaction edges, spawn inheritance deep copy.
- [ ] Integration: budget exceed path (soft warn 80%, stop 100%), JSON stdout purity regression test.
- [ ] Golden: default prompt resolution precedence order.
Docs
- [ ] CONTRIBUTING: architecture layers, adding a tool/provider, testing matrix, style/lint, release process.
- [ ] Memory architecture diagram (conversation vs vector vs shared store namespaces).
Security / Safety
- [ ] Shell tool allowlist & destructive command confirmation gate (config toggle).
- [ ] Network tool response size + content‑type limits.
- [ ] Redact secrets (heuristics) in debug/traces.

Legend: [x] done, [ ] pending. Purely internal items stay here until shipped; user‑visible features go to roadmap.

## Active / Near‑Term User‑Visible Work
High Priority
1. Context pipeline v2 (relevance + token budgeting) – replaces minimal builder.
2. Agent TODO tool (persistent memstore CRUD) – planning aid.
3. Delegated agent tool call dedupe per iteration.
4. Elapsed time counter per agent in TUI.
5. Startup diagnostics banner (prompt presence + API keys).

Medium
- Modern spinner + unified streaming tail indicator.
- Code syntax highlighting in TUI.
- Agent cycling key rebind (free arrow keys for cursor).
- Nerd Font optional glyph set with graceful fallback.

Deferred / Exploratory
- Adaptive context pack weighting & incremental pruning.
- Inline diff preview before patch apply.
- Advanced relevance scoring (hybrid semantic + structural graph metrics).

## Roadmap (Milestone Outline)
M1 (Hardening Sprint) – Core cleanup + context pipeline scaffolding + env helper consolidation + metrics label fix.
M2 – TODO tool + delegation dedupe + elapsed timers + startup diagnostics; add tests & CONTRIBUTING.
M3 – File locks (opt‑in) + sandbox enforcement + status board tool/TUI panel.
M4 – LSP symbol/ref/definition providers + RelatedFiles hybrid scoring integration.
M5 – Extended agent interaction (ask_user), semantic memory search, optional serve mode.
Later – Remote/cluster spawn, streaming event bus clients, adaptive context weighting.

## Context Management v2 (Summary)
Goal: deterministic, budget‑aware assembly vs ad hoc injection.
Pipeline: Providers → Score/Budget → Assemble → Messages.
Core Providers (initial set): TaskSpec, History (with compaction), WorkspaceSummary, ActiveFile, RelatedFiles, GitDiff, TestFailures, RunOutput, Memory, Rules, LSPDefs.
Mechanics:
- Each provider returns (id, rawContent, estTokens, relevanceScore, truncationFn).
- Budget allocator sorts by (score * weight), fits greedily, applies truncation on overflow.
- Always annotate with provenance headers (e.g. `<<file:path.go:23-57>>`).
Testing: pack ordering, truncation boundary, provider opt‑out when empty.

## Agent TODO Tool (Spec Snapshot)
Namespace: `todo:project:<project_path_hash>`.
Item fields: id, title, description, priority (low|medium|high), tags[], agent_id, status (pending|done), created_at, updated_at.
APIs: todo_add, todo_list (filters: status, tags, agent_id, limit), todo_update, todo_delete, todo_get.
TTL: optional cleanup for done items (configurable env/role).
Tests: CRUD happy path, tag filter, TTL expiry (simulated), invalid id update/delete.

## Current Command Line Usage
### Direct Prompt Invocation
You can invoke agentry with a direct prompt:
```bash
./agentry "your prompt here"
```

Useful for:
- Delegation tests
- One‑offs / CI
- Debugging

Example:
```bash
./agentry "Review PRODUCT.md and give a concise bullet summary (delegate only if needed)"
```

### Available Commands
```bash
# Start TUI (default when no command provided)
./agentry
./agentry tui

# Direct prompt execution  
./agentry "create a hello world program"

# Command utilities
./agentry analyze <trace-file>          # Analyze trace files / (cost integrated)
./agentry refresh-models                # Download latest model pricing
./agentry invoke ...                    # One-shot JSON call
./agentry team ...                      # Team operations (roles, spawn, call)
./agentry memory ...                    # SharedStore ops
./agentry version                       # Show version
./agentry help                          # Show help

# Deprecated (shows warning; removal pending)
./agentry chat
./agentry dev
```

### Common Flags (TUI Mode)
```bash
./agentry --config path/to/.agentry.yaml
./agentry --theme dark
./agentry --save-id session1
./agentry --resume-id session1
```

### Debug Mode
```bash
AGENTRY_DEBUG=1 ./agentry "test prompt"   # Debug output (stderr / log)
AGENTRY_DEBUG=1 ./agentry                 # TUI debug (logs redirected)
```

Note: TUI mode redirects verbose debug output to avoid UI interference.

## Tightening Review (Key Current Issues)
Identified areas (do before adding major new features):
1. Monolithic `Agent.Run` – extract tool execution & streaming; add ctx cancellation checks.
2. Duplicated env helper functions – unify; improves testability.
3. High cardinality metrics (UUID labels) – replace with role labels.
 4. Missing fallback system prompt – protect against blank system context. (DONE)
5. Manual O(n^2) sort in history compaction – replace with `sort.Slice`.
6. Verbose debug always printing system prompt/tool names – guard behind selective debug levels.
7. Inconsistent naming: legacy platform injection helper; unified naming & deprecate legacy wrapper.
8. Lack of explicit model options struct – future unification, though not user visible now.
9. Potential double counting (already fixed for tool output tokens; keep regression test).
10. Absence of cancellation/timeouts around external tool execution – add context checks.
11. Potential secret leakage in traces (env prints) – redact patterns.
12. Shared vector store semantics undocumented – add inline comment + README clarification.

## Update Policy
After material change: update this file + relevant docs + role templates (if affected). Keep backlog accurate (remove shipped, avoid stale duplicates).

## Status Legend
Internal Hardening tasks stay until merged; user roadmap items move to Completed once minimal slice is merged & documented.

---
Historical separate PLAN.md & FEATURES.md merged here (2025‑08‑18).
---


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
./agentry "Review PRODUCT.md and give a concise bullet summary (delegate only if needed)"
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
### Current Status Snapshot
Core:
- Minimal context builder in place; heavy hardcoded context removed (now using sentinel + cached project summaries).
- Delegation safety: worker agents stripped of `agent` tool to avoid recursive cascades.
- Token & cost tracking integrated with live progress display in TUI.

TUI UX:
- Input history (Up/Down) implemented.
- Conditional autoscroll prevents fighting the stream when scrolled up.
- Basic spinner still legacy (| / - \) – modernization pending.

Planned (High Priority)
1. Context pipeline (Provider → Budget → Assembler) with relevance + token budgeting.
2. Agent TODO list tool (persistent in memstore) with add/list/update/delete/get APIs.
3. Delegated agent call deduplication (single `agent` tool execution per iteration) to reduce redundant spawns.
4. Elapsed time counter per active agent in sidebar.
5. Startup diagnostics banner (prompt file + API key presence) beneath logo.

Planned (Medium)
- Modern mini-dots spinner + unified streaming tail indicator.
- Code syntax highlighting in TUI.
- Shift+Left/Right (or alternative) for agent cycling; reserve Left/Right for cursor.
- Nerd Font glyph optional support with graceful fallback.

Deferred / Exploratory
- Rich context pack weighting heuristics & adaptive pruning.
- Inline diff preview before patch apply.

### Immediate Next Steps
1. Implement context provider registry & budgeting harness.
2. Ship TODO tool (schema already sketched) storing under `todo:project:<path>`.
3. Add single-iteration agent tool deduper (light filter in `Agent.Run`).
4. Introduce elapsed timer + refactor spinner frames.
5. Backfill tests covering: context truncation, TODO CRUD, delegation dedupe.

Track progress in `FEATURES.md` (now separated into Completed / Planned sections).
