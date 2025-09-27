# Agentry Product Brief

_Last updated: September 2025_

Agentry is a local-first agent runtime with a built-in TUI, structured tracing, and a curated set of tools for day-to-day development automation.

## Scope

- **Runtime:** streaming loop, tool execution, budgeting, error handling
- **TUI:** conversational view, tool output log, TODO board, token/cost footer
- **Configuration:** single `.agentry.yaml` drives models, tools, and role templates
- **Tooling:** safe-by-default builtin tools for files, search, shell, networking, and delegation
- **Observability:** JSONL traces, rolling debug logs, model pricing cache

## Vision

Deliver a dependable assistant that can plan and execute development tasks in a single repository without cloud dependencies. Keep the footprint small, make behaviour observable, and prioritise quality over breadth.

## What Exists Today

- Core conversation loop with streaming responses and structured tracing
- Cost manager backed by a refreshable pricing cache
- Delegation via the `agent` tool and SOP-backed role templates
- Builtin TODO store surfaced in the TUI
- Debug wrappers (`scripts/debug-agentry.sh`) and logging smoke tests
- Go 1.25 feature adoption where it improves concurrency/readability

## Near-Term Priorities

1. Ship the Context-Lite prompt compiler with golden tests
2. Expand unit coverage for tool execution error paths and budgeting
3. Polish the TUI TODO board (filters, clearer status updates)
4. Refresh SOP/role templates to match the trimmed tool set
5. Keep documentation in sync with the simplified CLI/TUI surface

## Recently Completed

- Removed legacy features: persistent sessions, NATS queues, Kubernetes deployment, eval system, examples folder
- Simplified CLI to `tui`, direct prompt execution, and `refresh-models`
- Relocated helper scripts/log docs into `scripts/` and `docs/`
- Added regression tests for streaming aggregation and tool executor behaviour

## Principles

- Minimal dependencies and fast startup
- Everything observable: traces, debug logs, cost summaries
- Config-first; no hidden magic outside `.agentry.yaml`
- Documentation and tests updated alongside every feature

Keep this document terse and current—update it whenever priorities change.
