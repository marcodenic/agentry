# Test Coverage Roadmap

## Configuration & Environment
- [x] `internal/config`: cover loader behaviors (env overrides, flag inputs, deprecated fields)
- [x] `internal/env`: add tests around variable resolution and error reporting
- [x] `cmd/agentry`: smoke-test CLI flag parsing and config wiring

## Core Agent Workflow
- [x] `internal/core`: exercise conversation lifecycle (preparation, streaming, post-processing)
- [x] `internal/contracts`: ensure interfaces stay backward-compatible via compile-time assertions or minimal usage tests
- [x] `internal/cost`: verify pricing calculators with sample transcripts

## Team Coordination
- [x] `internal/team`: add unit tests for delegation sessions, coordination history, and shared state mutations
- [x] `internal/teamruntime`: cover telemetry helpers and notifier integration
- [x] `internal/teamruntime/testsupport`: document intent and add smoke tests for fixtures (or collapse into main package if redundant)
- [x] `internal/sop`: cover registry filtering, condition evaluation, and prompt formatting

## Tooling & Patch Helpers
- [x] `internal/tool`: expand beyond existing tests to include delegation/builtins edge cases
- [x] `internal/patch`: verify patch application helpers against typical edit scenarios
- [x] `internal/tokens`: add regression tests for token counting utilities
- [x] `internal/lsp`: exercise cross-language diagnostic parsing and execution gating

## UI & Memory Layers
- [x] `internal/tui`: increase coverage for key models/view logic beyond integration suite
- [x] `internal/memory`: add tests for shared memory lifecycle helpers
- [x] `internal/statusbar` and `internal/glyphs`: snapshot tests to prevent accidental regressions

## Observability & Misc
- [x] `internal/audit`: cover file rotation and error handling
- [x] `internal/debug`: cover logging toggles and file sink wiring
- [x] `internal/trace`: add tests for trace event formatting
- [x] `internal/version`: ensure release constant stays semver-aligned
- [x] `tests`: ensured integration suite runs in sandbox (added deterministic agent clients, network skips, and regression coverage for delegation flows)

## Tracking & Next Steps
- [x] Prioritize high-risk packages (`internal/config`, `internal/core`, `internal/team`) and file detailed test charters (see below)
- [ ] Schedule incremental PRs to land coverage, updating this roadmap after each batch ships
- [ ] Introduce reusable test agent factories so delegated roles don't fall back to `model.NewMock` and require manual injection
- [ ] Establish coverage targets (package-level percentages and critical-path scenarios) once initial suites exist

### High-Risk Package Charters

**internal/config**
- Exercise `AGENTRY_CONFIG_HOME` fallback to `$HOME/.config/agentry`, including unreadable global configs and empty overlays.
- Add matrix of include/merge scenarios: project-level include list appended vs. overridden, vector store partial updates, zero-value budgets.
- Validate env-driven overrides once implemented (currently only TODO comment) to lock expected precedence before wiring code.

**internal/core**
- Drive streaming aggregator through: empty channel error, chunk errors mid-stream, zero token counts forcing `tokens.Count` fallback, and response linking (non-empty `responseID`).
- Cover `handleCompletion` cost gating (`AGENTRY_STOP_ON_BUDGET` true/false), memory append skipping, and plan-follow-up rate limits when heuristics disabled.
- Probe tool executor branches: JSON validation failures, `TreatErrorsAsResults` toggle, notifier selection (`AGENTRY_TUI_MODE`), and terminal tool exit short-circuit.

**internal/team**
- Integration-style delegation session: spawning missing agent, timeout override via `AGENTRY_DELEGATION_TIMEOUT`, workspace context injection, and shared-memory bookkeeping on success vs. timeout-with-work.
- Concurrency tests for task orchestration (`AssignTask`, `ExecuteParallelTasks`) to surface current unsynchronized status updates; decide between locks or channel handoff.
- Codify expectations for coordination helpers (history trim, `PublishWorkspaceEvent`) and coverage around `checkWorkCompleted` heuristics to avoid filesystem thrash.

### Dead / Disconnected Code Candidates
- `internal/team/help.go:ProposeCollaboration` is exported yet unreferenced; either wire into tooling or drop from surface area.
- `internal/team/coordination.go` and `coordination_go125.go` mutate `Task` fields in goroutines without synchronization, inviting data races during `ListTasks` / `GetTask`; consider centralizing task lifecycle updates behind a lock or message queue.
- `internal/config/Load` still advertises environment overrides but never applies them; clarify intent (implement or remove comment) to avoid false confidence during audits.
