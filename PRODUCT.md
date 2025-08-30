# Agentry Product & Roadmap

Single authoritative doc. Keep terse, actionable. Update after each merge/re-prioritization.

FOR AGENTS, run: ./agentry <prompt>

## Vision (Condensed)

Local-first, observable, resilient **multi-agent** development orchestrator. Open any repo, point Agent 0 at a planning doc or task list, and it **delegates → implements → tests → reviews → finalizes (PR/commit)** with clear traces and accurate cost accounting.

## Current Foundations (What Exists)

* **Core loop:** tool calling, streaming, tracing, cost/token accounting (per turn), error-as-data, retry caps.
* **Tools:** 30+ built-ins (atomic file ops, search/replace, web/network, OpenAPI/MCP, audit/patch, delegation/spawn).
* **Models:** OpenAI + Anthropic via unified `model.Client` (streaming; usage tracked).
* **Multi-agent:** Team registry + delegation; Agent 0 role = orchestrator (spawn/manage workers).
* **Memory:** per-agent convo history + vector store; SharedStore (mem/file) for persistence; **basic checkpointing** (low priority to extend).
* **TUI:** live stream, delegation events, token/cost bar, diagnostics summary, safe autoscroll.
* **CLI:** Direct prompts (+ default TUI when no command) and `refresh-models`. Legacy commands (`invoke`, `team`, `memory`, `analyze`) removed.
* **Context:** minimal builder (v2 pipeline incoming). Platform/roles injected via files.

## Recently Completed (Highlights)

* SharedStore (mem+file) with TTL/GC; persist coordination events.
* Inbox tools: `inbox_read`, `inbox_clear`, `request_help`, `workspace_events` (+ auto-inject unread).
* Delegation safety: worker agents lose `agent` tool.
* LSP diagnostics surfaced in TUI (gopls / tsc).
* Minimal context builder shipped; heavy hardcoded text removed.
* Pricing cache path moved to user cache dir; `refresh-models` command available.
* Iteration cap removed (agent runs to final) with optional budget stop.
* **Sprint Complete: Major Cleanup & Simplification**
  * Removed Prometheus/metrics system completely (code, deps, configs)
  * Removed eval system entirely (internal/eval, commands, test files)
* Simplified CLI to core functionality: direct prompts (default TUI with no command) and `refresh-models`
* Eliminated legacy commands: invoke, team, memory, cost, analyze
  * Agent 0 TUI display fixed (now shows "Agent 0" with "System" role)
  * All tests passing; cleaner, lighter codebase ready for Context v2

---

## Hardening & Cleanup (No New UX; ship fast)

Architecture

* [x] **Context pipeline**: Provider → Budget → Assembler.
* [ ] Introduce `AgentConfig` (budgets, error handling, model name) to reduce env sprawl.
* [x] Extract tool execution from `Agent.Run` → `executeToolCalls` (smaller CC, testable).
* [x] Cancellation checks pre/post model call & per tool.
* [x] Fallback minimal system prompt if role file missing.

Code Quality

* [x] Consolidate env helpers into `internal/env`.
* [x] Replace O(n²) compaction sort with `sort.Slice`.
* [ ] Clarify spawn semantics (shared vs isolated vector store) + option toggle.
* [ ] Collapse duplicate default prompt helpers; keep one public API.
* [ ] Guard verbose dumps behind selective debug channels.
* [ ] Normalize model names (`provider/model`) via helper.

**Remove/Retire (this sprint)**

* [x] **Remove Prometheus/metrics** code & deps; delete metric labels/counters/histograms.
* [x] Remove legacy **eval** paths/flags/commands and dead modes.
* [x] **Remove legacy CLI commands** (invoke, team, memory, cost, analyze) - simplified to direct prompts.
* [x] Purge unused metrics env vars & docs.


Testing

* [ ] Unit: tool error recovery (consecutive cap), history compaction edges, spawn inheritance deep copy.
* [ ] Integration: budget warnings (80%/100%), JSON stdout purity regression.
* [ ] Golden: prompt resolution precedence.

Docs

* [ ] CONTRIBUTING: layers, adding tool/provider, test matrix, release steps.
* [ ] Memory architecture diagram (conversation vs vector vs shared store).

---

## Priorities (User-Visible)

### High

1. **Context Pipeline v2** (relevance + token budgeting; see spec below).
2. **Agent TODO & Planning Memory**

   * Persistent TODOs & project planning memory (CRUD, filters, TTL); planning-doc ingestion.
3. **Cost Accuracy Pass**

   * Correct **input+output** token accounting per model; pricing from `models.dev` table; per-agent/session totals; budgets.
4. **AST-Based Editing v1**

   * Surgical, syntax-aware edits (rename, insert/import, replace by query) for Go/TS/JS (Python optional); graceful text fallback.
5. **Auto-LSP Loop**

   * Run diagnostics automatically after edits; feed results into next turn; TUI summary; fix-it guidance.
6. **Agent 0 Orchestration Loop**

   * Explicit “plan → delegate → build → test → review/critic → integrate/PR → verify-done” loop in role & runtime hooks.

### Medium

* TUI polish: modern spinner, unified stream tail, code syntax highlighting.
* Agent cycling keybind fix; Nerd Font optional glyphs.
* Inline diff preview before patch apply (UX).

### Deferred / Later

* Advanced checkpointing; resumable workflows UX.
* Multi-model/provider plugins (beyond OpenAI/Anthropic).
* Remote/cluster spawn; event bus; distributed teams.
* Guardrail frameworks; sandbox hardening.

---

## Roadmap (Milestones)

**M1 — Hardening & Cost** (In Progress)

* [x] Remove Prometheus/metrics + eval code.
* [x] CLI simplification (direct prompts, removed legacy commands).
* [ ] Context v2 scaffolding + token counters.
* [ ] Cost correctness (input/output/tool) + pricing loader + TUI totals/budgets.
* [ ] OpenAI/Anthropic usage normalization (streaming + usage unify).

**M2 — Memory & Orchestration**

* Ship **TODO tool** + project planning memory.
* Planning-doc provider (detect TASKS.md/PROJECT.md/ROADMAP.md).
* Agent 0 loop hooks: completion verification + critic pass; success criteria.

**M3 — LSP & Tests Loop**

* Auto-LSP diagnostics post-edit; TUI diag panel.
* Test-runner integration (detect `go test` / `npm test` / `pytest`) and auto-run after builds; feed failures back.

**M4 — AST Editing v1**

* Tree-sitter or language parsers for Go/TS/JS: `rename_symbol`, `replace_by_query`, `ensure_import`, `apply_patch_tree`.
* Fallback to line edits when AST fails; validate with formatter/linter.

**M5 — UX Polish**

* Diff preview, syntax highlight, spinner/stream improvements.
* Docs: CONTRIBUTING + memory architecture.

---

## Context Management v2 (Summary)

Deterministic, budget-aware assembly.

**Pipeline:** Providers → Score/Budget → Assemble → Messages
**Core Providers (initial):**

* `TaskSpec` (user ask + agent role)
* `PlanningDoc` (extract tasks/goals from TASKS.md/PROJECT.md/ROADMAP.md)
* `History` (compacted, with auto-summary on overflow)
* `WorkspaceSummary` (tree outline)
* `ActiveFile` (windowed excerpt)
* `RelatedFiles` (hybrid scoring: lexical + semantic + structural)
* `GitDiff` (staged/unstaged)
* `TestFailures` (recent failing traces)
* `RunOutput` (last command output)
* `LSPDefs` (defs/refs/hover snippets)
* `Memory` (project KB/TODOs)

**Budgeting:**

* Compute available tokens (`contextLimit(model) - system - user`).
* Rank by `score * weight`; include greedily; apply truncation (prefix/suffix/outline).
* Always annotate provenance: `<<file:path.go:23-57>>`.

---

## Agent TODO & Planning Memory (Spec)

**Namespace:** `todo:project:<hash>`
**Item:** `id, title, description, priority(low|med|high), tags[], agent_id, status(pending|done), created_at, updated_at`
**APIs:** `todo_add`, `todo_list(filters)`, `todo_update`, `todo_delete`, `todo_get`
**PlanningDoc ingestion:** parse structured bullets/checkboxes into TODOs; link back to source (file\:line).
**Persistence:** SharedStore (file or sqlite); optional TTL for done.

---

## Cost Accuracy (Spec)

**Goals:** exact token & cost visibility; budget enforcement.

* **Usage accounting:** count **prompt (system+user+tool args)** and **completion** tokens; include tool-returned function messages if billable; handle streaming chunks cumulatively.
* **Pricing source:** load from `models.dev` mapping into `models_pricing.json` (cached by `refresh-models`).
* **Displays:** per-turn + per-agent + session totals in TUI; cost per tool; first-token latency retained (as a trace field, not metrics).
* **Budgets:** soft warn at 80%, hard stop at 100% (env configurable).
* **Tests:** golden tests for usage parsing; cross-provider parity (OpenAI/Anthropic).

---

## AST-Based Editing v1 (Spec)

**Languages:** Go, TypeScript/JavaScript (Python optional).
**Backends:** tree-sitter or native parsers (`go/ast`, `ts-morph` via child proc).
**Operations:**

* `rename_symbol(file, from, to, scope?)`
* `replace_by_query(file, query, replacement)` (TSQ/AST query)
* `ensure_import(file, module, name?)`
* `apply_patch_tree(file, patchSpec)` (insert/move/delete nodes)
  **Validation:** run formatter (`gofmt`, `eslint --fix`/`prettier`) and basic build after each edit; auto-LSP diagnostics.
  **Fallback:** degrade to line-level `edit_range` with syntax check if AST parse fails.

---

## Auto-LSP Loop (Spec)

* Detect language → start server(s) as needed; cache handles per workspace.
* **Trigger:** after any file write/patch, run diagnostics automatically; batch per tick to avoid thrash.
* **Surface:** TUI panel (errors/warnings per file), quick links, summarized hints.
* **Feed back:** append diagnostics into next agent turn context (`LSPDefs`, `TestFailures`).
* **Cache:** defs/refs/hover by file+version to reduce redundant queries.

---

## Agent 0 Orchestration Loop (Runtime + Role Addendum)

**Loop skeleton (implicit; no external DSL):**

1. **Plan** (read PlanningDoc/TODOs; create/update TODOs).
2. **Delegate** (spawn coder/tester/reviewer agents as needed).
3. **Build** (coder edits; AST where possible).
4. **Test** (auto detect & run project tests/build; capture failures).
5. **Review/Critic** (spawn critic/reviewer; verify acceptance criteria).
6. **Integrate** (commit/PR when green; reference tasks).
7. **Verify-Done** (re-run tests; sanity checks; close TODOs).
8. If not done → **iterate** with bounded retries/budget.

**Role template updates:** explicit success criteria; require critic pass; require zero diagnostics in touched files, or acknowledged waivers.

---

## CLI Usage

Agentry has been simplified to focus on core functionality. The CLI supports four main commands:

### Primary Usage

```bash
# Direct prompt execution (Agent 0 responds directly)
# Quotes are optional for simple prompts:
agentry hello there
agentry fix the failing tests
agentry implement a health check endpoint

# Use quotes when you need special characters or want to be explicit:
agentry "analyze the codebase structure & suggest improvements"
agentry "implement user auth with JWT tokens"

# Start TUI interface (default when no command)
agentry
agentry --config custom.yaml

# Update model pricing data
agentry refresh-models

# Show help
agentry help
```

### Common Flags

All commands support these flags:

```bash
--config path/to/.agentry.yaml    # Config file path
--theme dark|light|auto           # Theme override  
--debug                          # Enable debug output
--keybinds path/to/keybinds.json # Custom keybindings
--creds path/to/creds.json       # Credentials file
--mcp server1,server2            # MCP servers list
--save-id session1               # Save conversation state
--resume-id session1             # Resume from saved state
--checkpoint-id ckpt1            # Checkpoint session ID
--port 8080                      # HTTP server port
```

### Examples

```bash
# Simple tasks (no quotes needed!)
agentry list all TODO comments in the codebase
agentry fix the failing tests
agentry add error handling to the auth module

# With flags for debugging
agentry --debug analyze the codebase structure

# TUI with session resumption
agentry --resume-id my-session --theme light

# Complex prompts (quotes recommended for special characters)
agentry "implement user auth with JWT tokens & refresh logic"
agentry "analyze performance bottlenecks using profiling data"

# Update models
agentry refresh-models
```

### Environment Variables (Alternative Configuration)

While flags are preferred, these environment variables are still supported:

**Core Settings:**
- `AGENTRY_DEBUG=1` - Enable debug output (use `--debug` flag instead)
- `AGENTRY_THEME=dark` - Set theme (use `--theme` flag instead)
- `AGENTRY_ENV_FILE=/path/to/.env` - Load environment from file

**Advanced/Internal Settings:**
- `AGENTRY_TUI_MODE=1` - Internal flag (set automatically)
- `AGENTRY_DEFAULT_PROMPT="..."` - Default prompt override
- `AGENTRY_AUDIT_LOG=/path/to/audit.log` - Audit logging
- `AGENTRY_HISTORY_LIMIT=100` - Chat history limit
- `AGENTRY_DELEGATION_TIMEOUT=300` - Agent delegation timeout
- `AGENTRY_MODELS_CACHE=/path/to/cache` - Model cache location
- `AGENTRY_STORE_GC_SEC=3600` - Memory store garbage collection interval

**Tool/Filter Controls (Deprecated - Use CLI flags instead):**
- ~~`AGENTRY_DISABLE_TOOL_FILTER=1`~~ - Use `--disable-tools` flag
- ~~`AGENTRY_TOOL_ALLOW_EXTRA=tool1,tool2`~~ - Use `--allow-tools` flag  
- ~~`AGENTRY_TOOL_DENY=tool1,tool2`~~ - Use `--deny-tools` flag
- `AGENTRY_DISABLE_CONTEXT=1` - Disable context pipeline

**Context/Memory Tuning:**
- `AGENTRY_CTX_CAP_AGENT0=8000` - Context limit for Agent 0
- `AGENTRY_CTX_CAP_WORKER=4000` - Context limit for worker agents

**Logging/Communication:**
- `AGENTRY_COMM_LOG=1` - Enable communication logging
- `AGENTRY_COLLECTOR=...` - Telemetry collector endpoint
- `AGENTRY_PORT=8080` - HTTP server port (use `--port` flag instead)

### Migration from Environment Variables

Many environment variables can be replaced with cleaner CLI flags:

```bash
# Old way
AGENTRY_DEBUG=1 AGENTRY_THEME=dark agentry "task"

# New way  
agentry --debug --theme dark "task"
```

For advanced users who need environment variable control for automation or CI/CD, they remain supported, but flags are recommended for interactive use.

---

## Next Steps (Tight List)

1. [x] **Rip metrics & eval**: remove code/deps/flags/docs.
2. **Context v2 harness** + token counters + provider stubs (TaskSpec, PlanningDoc, History, ActiveFile, RelatedFiles, LSPDefs, GitDiff, TestFailures, RunOutput, Memory).
3. **TODO tool** + PlanningDoc ingestion + persistence.
4. **Cost pass**: accurate usage parsing, pricing loader, TUI totals + budgets.
5. **Auto-LSP** post-edit loop + TUI diagnostics.
6. **AST v1** ops (Go/TS/JS) + formatter/diag validation + fallback.
7. **Agent 0 loop** addendum in role + runtime hooks (critic + success checks).
8. **Tests**: context truncation, TODO CRUD, cost accounting, diag loop, AST operations.

---
## BUGS

- on resizing the window we get like: marlformed char codes or something.
- no reasoning_effort support



**Update Policy:** After material change, update this file + role templates + CLI help. Keep backlog clean (remove shipped; no stale dupes).

**Status Legend:** Internal hardening stays until merged; user-visible items move to “Recently Completed” once the minimal slice is shipped & documented.

*Historical PLAN.md & FEATURES.md merged here (updated 2025-08-18).*
