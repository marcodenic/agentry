# Delegated Agent Visibility Plan

## Goals
- Surface delegated agent activity alongside Agent 0 so operators stay informed about background work.
- Present live tool usage (name + relevant arguments) in a compact, modern Bubble Tea component.
- Preserve existing Agent 0 flow while introducing a scalable layout that handles multiple concurrent workers.

## 1. Multi-Agent State Tracking (`internal/tui/model.go`, `internal/team/delegation_session.go`)
- [ ] Add delegation hooks that emit spawn/teardown events (e.g., from `newDelegationSession` in `internal/team/delegation_session.go:26`) so the TUI knows when workers start/finish.
- [ ] Extend the TUI model to append `AgentInfo` entries when new agent IDs appear, including spinner/token progress initialization (`internal/tui/model.go:107`).
- [ ] Ensure trace pipes are opened for delegated agents so `trace.EventToolStart` / `EventToolEnd` fire for every worker (`internal/tui/runtime_helpers.go:322`).

## 2. Shared Activity Feed Component (`internal/tui/components/activity_feed.go` new)
- [ ] Create a Bubble Tea list/table component that renders tool actions with columns for time, agent, tool, and summarized args (derived from `formatToolAction` in `internal/tui/runtime_helpers.go:14`).
- [ ] Persist the last N events (per agent + global feed) to support quick scrolling and filtering.
- [ ] Add styling consistent with existing gradient/branding (reuse lipgloss palette and glyphs from `internal/tui/theme.go` and `internal/tui/runtime_helpers.go`).

## 3. Tool Telemetry & Args Capture (`internal/team/events.go`, `internal/tui/runtime_helpers.go`)
- [ ] Emit workspace/coordination events even when `AGENTRY_TUI_MODE=1` so delegated agent tool usage reaches the front-end (`internal/team/events.go:70`).
- [ ] Normalize argument payloads (convert JSON → map) before sending to the TUI to avoid ad-hoc string formats.
- [ ] Update `formatToolAction` to handle multi-agent context: include agent badge, highlight destructive commands, and gracefully truncate long args.

## 4. Layout & Navigation (`internal/tui/model.go`, `internal/tui/view.go`)
- [ ] Introduce a split-pane layout: left = Agent 0 stream, right = activity feed / delegated agent tabs (Bubble Tea `viewport` + `list`).
- [ ] Provide keybinds to focus the activity feed, jump between agents, and collapse/expand delegated panels (extend `Model.keys` + `model_keys.go`).
- [ ] Display per-agent headers with status badges (idle/working/error) and last tool invocation underneath.

## 5. Testing & Telemetry (`internal/tui`, `tests`)
- [ ] Add smoke tests that simulate trace/tool events for two agents and assert feed updates (can leverage go test with fake trace emitter).
- [ ] Validate performance with many events; add configurable cap + GC for activity entries.
- [ ] Document new UX in `docs/tui.md` (screenshots + usage) once implemented.
