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

## Tooling & Patch Helpers
- [x] `internal/tool`: expand beyond existing tests to include delegation/builtins edge cases
- [x] `internal/patch`: verify patch application helpers against typical edit scenarios
- [x] `internal/tokens`: add regression tests for token counting utilities

## UI & Memory Layers
- [x] `internal/tui`: increase coverage for key models/view logic beyond integration suite
- [x] `internal/memory`: add tests for shared memory lifecycle helpers
- [x] `internal/statusbar` and `internal/glyphs`: snapshot tests to prevent accidental regressions

## Observability & Misc
- [x] `internal/debug`: cover logging toggles and file sink wiring
- [x] `internal/trace`: add tests for trace event formatting
- [ ] `tests`: ensure end-to-end suites exercise new features as they land (add cases or harness extensions as needed)

## Tracking & Next Steps
- [ ] Prioritize high-risk packages (`internal/config`, `internal/core`, `internal/team`) and file detailed test charters
- [ ] Schedule incremental PRs to land coverage, updating this roadmap after each batch ships
- [ ] Establish coverage targets (package-level percentages and critical-path scenarios) once initial suites exist
