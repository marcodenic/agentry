# Refactor Opportunities

## Core Agent Loop (`internal/core/agent.go`)
- Extract a conversation lifecycle object that handles preparation, model invocation, streaming, and post-processing so `Agent.Run` stops being a 400-line monolith.
- Move verbose logging/telemetry into composable decorators so agents read core logic without wading through instrumentation concerns.

## Tool Execution (`internal/core/agent.go`, `executeToolCalls`)
- Split validation, execution, result-shaping, and user feedback into dedicated helpers or strategies to reduce branching and make each behaviour easier to test.
- Consider encapsulating stderr UX and retry policy so per-tool tweaks do not require editing the central loop.

## Team Delegation (`internal/team/delegation.go`)
- Introduce a `DelegationSession` (or similar) type that owns agent lookup/spawn, progress reporting, timeout handling, and telemetry, leaving `Team.Call` focused on orchestration.
- Centralise workspace-event enrichment and logging helpers to keep delegation flow legible to agents.

## Team Runtime Coordination (`internal/team/delegation.go`, `internal/team/tools.go`)
- Separate parallel delegation logic from sequential delegation so each path has a minimal surface area and clearer error handling.
- Extract shared logging utilities (emoji status updates, file logging) into a lightweight notifier to avoid repetition across delegation paths.

## Team Tool Registry (`internal/tool/builtins_team.go`)
- Replace large literal maps of tool specs with manifest-driven registration (per-domain subpackages or data files) so agents can add/inspect coordination tools without scrolling hundreds of lines.
- Encapsulate schema-building helpers (alias resolution, shared strings) to avoid duplication between `agent` and `parallel_agents` tools.

## File Discovery & Todo Tooling (`internal/tool/file_discovery_builtins.go`, `internal/tool/todo_builtins.go`)
- Split discovery/todo tool registration into smaller modules organised by capability (search, listing, todo CRUD) to keep each file manageable.
- Introduce shared schema builders and response helpers so new tools can reuse argument validation and output formatting.

## Model Client (`internal/model/openai.go`)
- Break request construction, HTTP transport, streaming parsing, and telemetry into distinct structs/functions to isolate Responses API quirks from the generic `Client` interface.
- Provide a reusable streaming reader that can be unit-tested without hitting the network, reducing the mental load for agents editing the client.

## CLI Entry Points (`cmd/agentry/common.go`)
- Move flag parsing into a reusable options package so each command consumes a typed config struct; keep environment mutation in a narrow layer for clarity.
- Encapsulate configuration filtering (allow/deny tools, themes, credentials) into composable helpers to minimise side effects during CLI startup.

## TUI Surface Area (`internal/tui/*.go`)
- Break the root model into smaller feature models (input, history, diagnostics, robot, todo board) so agents can navigate the Bubble Tea update cycle without scanning 500+ lines.
- Relocate cosmetic helpers (logo gradients, glyph styling) into dedicated files to keep the core update/render logic concise.

## Repository Hygiene
- Treat compiled binaries (`bin/agentry`, `agentry`, `agentry-optimized`) and tracked logs under `debug/` as build artifacts; exclude or relocate them so agents have smaller contexts and faster searches.

## Additional Consideration
- The legacy context-building pipeline appears unused; plan for formal removal to prevent token overuse and simplify the delegation/runtime surface.
