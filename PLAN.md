# Agentry Plan — Central, Living Roadmap

This plan is the single source of truth for Agentry’s architecture, priorities, and progress. Update it with every material change (features, tests, docs). Keep it concise and actionable.

Last updated: 2025-08-10 (LSP diagnostics: structured parsing + TUI surface, team built-ins wiring)

---

## Vision

Build a fully functional, minimal-yet-powerful multi-agent orchestration framework that:
- Works on any project in any directory (local-first, cross-platform).
- Lets a system agent spawn/coordinate specialist agents via tools/roles.
- Maintains consistent shared memory/history across agents and runs.
- Is safe (permissions/sandbox), observable (metrics/tracing), and resilient (resume/checkpoint).
- Can self-improve using its own tools and workflows.

## Current State (summary)

Solid foundations:
- Core: lean Go run loop with tool-calls, tracing, metrics, and cost tracking.
- Tools: 30+ built-ins incl. atomic, line-precise file ops (edit_range, insert_at, search_replace, view, fileinfo), shell tools, OpenAPI/MCP.
- Models: mock/OpenAI/Anthropic via a clean model.Client; token and cost accounting.
- Multi-agent: Team orchestration with an agent delegation tool; TUI/CLI can spawn and switch agents; role templates exist.
- Memory: per-agent in-memory conversation store; simple in-memory vector store.

New since last update:
- CLI automation surface added: `invoke`, `team`, `memory` (JSON-first).
- TUI remains the default (no-arg). `chat`/`dev` print deprecation notices.
- Agent 0 (system) model updated to gpt-5 via role/config templates.
- Role loader fixes; team wiring registers the `agent` tool consistently.

 Gaps to close:
- Shared memory/history is ad hoc (map in Team), not durable or namespaced.
- Persistent stores (file/sqlite) for sessions, coordination, and vector indexing not wired yet.
 - Team built-ins (coordination/shared memory) largely implemented; ensure agent tool registration is consistent everywhere.
- Workflow/DAG orchestration is designed (types present) but not executed end-to-end.
- File collaboration safety (locks/watch) and sandbox enforcement are not integrated.
 - LSP: only diagnostics via gopls/tsc; no symbol/completion/navigation yet.

## Guiding Architecture Decisions

- SharedStore first-class: a simple key/value + list + optional pub/sub API for shared/team/project memory.
  - Backends: in-memory (default), file (JSON), sqlite; optional redis later.
  - Namespacing: project/<abs-path>, session/<id>, team/<id>, agent/<id>.
- Durable Memory: conversation memory and coordination events persist via SharedStore, with TTL + GC.
- Team Orchestration: team methods consume SharedStore for shared state and coordination logs; built-ins call Team methods (no placeholders).
- Workflow Orchestrator (MVP): small DAG runner with retries/timeouts; persists state and resumes on restart.
- Collaborative File Manager: opt-in locks around file-edit tools; change notifications via an event bus.
- Safety: per-role tool permissions, audit, and opt-in sandbox for shell/tool execution.
- Observability: Prometheus + OTLP tracing already present, extend to team/workflow.

## Milestones and Deliverables

M1 — Shared Memory + Real Team Built-ins (1–2 weeks)
- SharedStore interface + in-memory backend; file/sqlite adapters stubbed.
- Replace Team.sharedMemory map with SharedStore (namespaced).
- Persist conversation steps and coordination events; add TTL/GC respect from config.
- Make builtins_team call real Team methods (send_message, shared_memory, coordination_status, agent tool via TeamFromContext).
- Centralize agent tool registration when creating a Team; ensure TUI/CLI paths do this. (Partially done: CLI & TUI register via Team)
- Tests: shared store get/set/list; delegation uses team-backed agent tool; persistence smoke.

M2 — Workflow MVP + File Locks + TUI visibility (1–2 weeks)
- Minimal DAG runner: steps with dependencies; store execution state in SharedStore; resume after crash.
- Integrate file locks in edit_range/insert_at (opt-in via config); basic watcher notifications.
- TUI: add “Team/Memory” pane showing agents, recent coordination events, and recent shared keys.
- TUI: show LSP diagnostics summary and quick-action to run `lsp_diagnostics`. (partial: summary panel + Ctrl+D quick action implemented)
- Tests: 3-step workflow with 1 dependency; lock conflict test; resume test.

M3 — Durable Stores + Sandbox + Status Board (2–4 weeks)
- SQLite/file SharedStore adapters and Qdrant/Faiss vector store wiring with namespaces.
- Sandbox isolation for shell/builtins; enforce least-privilege per role.
- Status board backed by real updates; expose via tool and TUI.
- Tests: persistence across process restarts; sandbox denial; status summarization.

M4 — LSP Enhancements (1–2 weeks)
- Generic LSP client scaffold (stdio JSON-RPC) for multi-language support.
- Tools for symbol search, references, and rename (where supported).
- Optional completion hook to assist coding agents.
- Tests: gopls symbol lookup, tsc errors parsed, rename dry-run.

M4 — Remote Execution & Scale (later)
- Remote/cluster agent spawn and queue; SharedStore coordination across nodes.
- Streaming event bus to clients (SSE/WebSocket) for live collaboration.

## Task Backlog (prioritized)

- [ ] Define SharedStore interface and in-memory implementation.
- [ ] Swap Team.sharedMemory to SharedStore with namespaces (team/session/project).
- [ ] Persist conversation steps and coordination events; implement TTL/GC worker.
- [ ] Built-ins (team): implement real send_message, shared_memory (get/set/list), coordination_status using Team + SharedStore.
 - [x] Built-ins (team): implement real send_message, shared_memory (get/set/list), coordination_status using Team (SharedStore durability pending).
- [ ] Centralize agent tool registration when Team is created; ensure all entrypoints pass Team via context.
- [ ] Workflow MVP: execute steps -> Team.Call; persist step status/results; resume logic.
- [ ] File locks: integrate CollaborativeFileManager in edit_range/insert_at (config flag), with notifications.
- [ ] TUI: Team/Memory pane (agents, events, shared keys), and a quick “/team status” command.
- [ ] LSP: document and expose `lsp_diagnostics` in configs and README.
- [x] LSP: parse structured diagnostics (file:line:col:code:msg) for gopls/tsc.
- [x] LSP: TUI surface quick action and summary; highlight/jump-to-file hints deferred.
- [ ] Durable adapters: file/sqlite SharedStore; wire vector store backends with namespaces.
- [ ] Sandbox enforcement for shell/privileged tools; extend permissions per role.
- [ ] Status board updates + tool to fetch summary.
- [ ] E2E tests: workflow happy/failure, persistence/resume, file lock conflicts, sandbox denials.

Owners (initial)
- Architecture/store: @marcodenic
- Tools/team built-ins: @marcodenic
- TUI enhancements: @marcodenic

## Risks and Mitigations
- Concurrency: use RWMutex and non-blocking channels; cap event buffers; document drop policy.
- Data growth: cap shared memory values by size; prefer file references; TTL + GC.
- Deadlocks in locks: add lock expirations and stale-lock recovery.
- Safety: default-deny dangerous tools unless explicitly permitted in role config.
- Resume correctness: version workflow state; idempotent step execution.

## Quality Gates
- Build + Lint + Unit tests must pass on all supported platforms (Linux/macOS/Windows).
- Add integration tests for team delegation, shared memory, workflow, and file locks.
- Trace + metrics verified for key paths (agent runs, tool exec, workflow steps).

Status for this change set
- Build: PASS locally on Go 1.23.8; CLI builds; TUI compiles.
- Lint/Typecheck: PASS at compile-time; no new type issues observed.
- Unit tests: targeted LSP diagnostics test PASS; some unrelated legacy tests still failing (file ops/web tools) to be triaged separately.
- Smoke: TUI shows Diagnostics panel; Ctrl+D triggers lsp_diagnostics and renders counts; team built-ins operate against real Team via context.

## Update Policy
- After any substantive change: update PLAN.md, relevant docs in docs/, and (if applicable) role templates and examples. Keep this file authoritative and terse.

## Mode Consolidation (DX Contract)

Goal: minimize user-facing modes while keeping the TUI as the primary interactive UI and exposing a single, reliable automation surface for scripting and tests.

Decision (2025-08-10)
- Default behavior: running `agentry` with no subcommand launches the TUI.
- `chat` and `dev` are deprecated aliases to TUI and will be removed after a short deprecation window.

Single sources of truth
- Interactive: `agentry tui` (kept)
- Automation: headless CLI with JSON output (no interactivity). Optionally `serve` (HTTP/stdio) later.

Minimal command set (current/proposed)
- `tui` — interactive UI over the same engine.
- `invoke` — one-shot calls; supports `--agent` and `--trace` today; will add `--session`.
- `team` — `roles|list|spawn|call|stop` (all JSON by default).
- `memory` — `export|import` today; will add `get|set|list` with scopes.
- `workflow run --file examples/workflow.yaml` (supports `--json`).
- `serve` — optional HTTP/stdio API for external tools; JSON only.
- `analyze` — includes cost/tokens; deprecate separate `cost`.
- `version`.

Deprecations and aliases
- `chat` → alias to `tui` (deprecated; print guidance).
- `dev` → alias to `tui --debug` (deprecated; print guidance).
- Keep `pprof` under `analyze` or as a diagnostic subcommand; docs will reflect grouping.

Removal timeline
- Warn now via deprecation notices.
- Remove `chat`/`dev` in 2 minor releases (or after 2 weeks), updating docs and examples accordingly.

Global flags and behavior
- `--json` available on all non-TUI commands; disables prompts/ANSI, returns machine-readable results.
- `--config`, `--max-iter`, `--resume-id/--save-id/--checkpoint-id` honored consistently.

Implementation steps
1) Add `invoke`, `team`, `memory`, `workflow` subcommands backed by existing Team/Agent engine.
2) Add global `--json` and standardize JSON schemas for outputs and errors. (Partial: default JSON without flag)
3) Convert `cost` into `analyze --cost` and keep `pprof` under diagnostics.
4) Implement `chat`/`dev` aliases with deprecation warnings; update docs.
5) Ensure TUI uses the same engine (no exclusive codepaths); keep RegisterAgentTool wiring centralized.
6) Add VS Code tasks that call `invoke/team/workflow` with `--json` for easy in-editor automation.

## References
- README.md, docs/index.md, docs/usage.md, docs/CONFIG_GUIDE.md.
- Internal code: internal/core (agent loop), internal/team (orchestration), internal/tool (built-ins), internal/memory, internal/model.
- Example test project to drive orchestration: TEST_PROJECT.md.

---

Appendix: Example Driver Project
- TEST_PROJECT.md describes a bandwidth-monitor TUI. Use it as a realistic end-to-end scenario for multi-agent coordination (coder/tester/writer), file ops, shared memory, and workflow execution during development.
