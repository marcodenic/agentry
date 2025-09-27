# Refactor Opportunities (Next Wave)

## Core Conversation Pipeline (`internal/core/agent.go`, `internal/core/conversation.go`)
- [x] Extract a `PromptEnvelope`/`MessageBuilder` type so `buildMessages` stops interleaving logging, tool guidance, and prompt templating in one 150+ line block (`internal/core/agent.go:43`).
- [x] Split `conversationSession` into preparation, streaming, and post-processing components to make `loop` and `handleCompletion` testable without spinning the full agent (`internal/core/conversation.go:35`).
- [x] Introduce a reusable budget manager so the repeated `applyBudget` calls and token trimming logic live outside the agent struct, with focused unit tests for edge cases.

## TUI State & Layout (`internal/tui/model.go`, `internal/tui/runtime_helpers.go`)
- [x] Decompose the monolithic `Model` struct into feature-specific models (history, diagnostics, todo board, robot) so updates/renderers become composable (`internal/tui/model.go:18`).
- [x] Move the command/stream proxy helpers into a dedicated runtime service; `runtime_helpers.go` currently mixes IO plumbing with view mutations (`internal/tui/runtime_helpers.go:12`).
- [ ] Add integration-style tests for the split models to keep regressions in keyboard handling and viewport sizing under control.

## Team Coordination Surface (`internal/tool/team_tools.go`, `internal/team/*.go`)
- [x] Break `team_tools.go` into per-tool modules (status, shared memory, coordination events) that share a thin `teamServiceAdapter`; the file currently mixes schema definitions, formatting, and context lookups (`internal/tool/team_tools.go:18`).
- [ ] Consolidate repeated context extraction / logging in delegation helpers by introducing a `teamruntime` package that owns telemetry and notifier wiring (`internal/team/delegation_session.go:26`, `internal/team/logging.go`).
- [ ] Add tests around the adapter layer so each tool can be exercised without spinning an entire `Team`.

## TODO & Workspace Memory (`internal/tool/todo_builtins.go`, `internal/tui/memory.go`)
- [ ] Isolate TODO storage/serialization behind a small service (e.g., `internal/todo`); the current builtin mixes memstore reads, JSON, and command UX (`internal/tool/todo_builtins.go:18`).
- [ ] Reuse that service from the TUI memory board to remove duplicated key hashing and sorting logic (`internal/tui/memory.go`).
- [ ] Document/store schema migrations to avoid ad-hoc key evolution in future builtins.

## CLI Configuration & Flags (`cmd/agentry/common.go`, `internal/config`)
- [x] Replace the manual dual-flag binding (`max_iter`/`max-iter`, `http_timeout`/`http-timeout`) with a small alias-aware helper to retire the duplicated `IntVar` definitions (`cmd/agentry/common.go:32`).
- [x] Move side-effectful env tweaks (theme, audit log, disable context) behind a dedicated configuration mutator instead of writing to `os.Getenv` throughout `applyOverrides`.
- [x] Back these helpers with unit tests that confirm precedence rules (flags > env > config file).

## OpenAI Client Cohesion (`internal/model/openai.go`, `internal/model/openai_stream_reader.go`)
- [ ] Finish the streaming refactor by moving request construction, tracing, and reader wiring into a lightweight `OpenAIConversation` struct so `OpenAI.Stream` only orchestrates lifecycle (`internal/model/openai.go:150`).
- [ ] Fold the new `openAIStreamReader` into an interface so Anthropic/other clients can reuse the same contract, reducing future duplication.
- [ ] Expand the new reader tests to cover error events (`response.error`, malformed deltas) and deadline cancellations.
