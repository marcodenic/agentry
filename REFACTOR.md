# Test Coverage Roadmap

## Tracking & Next Steps
- [ ] Schedule incremental PRs to land coverage, updating this roadmap after each batch ships
- [ ] Introduce reusable test agent factories so delegated roles don't fall back to `model.NewMock` and require manual injection
- [ ] Establish coverage targets (package-level percentages and critical-path scenarios) once initial suites exist

## High-Risk Package Charters

**internal/config**
- Exercise `AGENTRY_CONFIG_HOME` fallback to `$HOME/.config/agentry`, including unreadable global configs and empty overlays.
- Add matrix of include/merge scenarios: project-level include list appended vs. overridden, vector store partial updates, zero-value budgets.
- Validate env-driven overrides once implemented (currently only TODO comment) to lock expected precedence before wiring code.

**internal/core**
- [ ] Add zero-token fallback coverage to the streaming aggregator tests.
- [ ] Verify notifier selection, JSON validation failure handling, and terminal tool exit short-circuit paths.

**internal/team**
- Run integration-style delegation session coverage: spawning missing agent, timeout override via `AGENTRY_DELEGATION_TIMEOUT`, workspace context injection, and shared-memory bookkeeping on success vs. timeout-with-work.
- Add concurrency tests for task orchestration (`AssignTask`, `ExecuteParallelTasks`) to surface current unsynchronized status updates; decide between locks or channel handoff.
- Codify expectations for coordination helpers (history trim, `PublishWorkspaceEvent`) and coverage around `checkWorkCompleted` heuristics to avoid filesystem thrash.

## Dead / Disconnected Code Candidates
- `internal/team/help.go:ProposeCollaboration` is exported yet unreferenced; either wire into tooling or drop from surface area.
- `internal/team/coordination.go` and `coordination_go125.go` mutate `Task` fields in goroutines without synchronization, inviting data races during `ListTasks` / `GetTask`; consider centralizing task lifecycle updates behind a lock or message queue.
- `internal/config/Load` still advertises environment overrides but never applies them; clarify intent (implement or remove comment) to avoid false confidence during audits.
- `internal/core/tool_exec.go` leaves `hadErrors` false when `TreatErrorsAsResults` is enabled; callers relying on that flag may miss soft failures—decide whether to increment or rename for clarity.
