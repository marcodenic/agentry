# Agentry Product Brief

_Last updated: December 2025_

Agentry is a local-first AI agent runtime with a built-in TUI, structured tracing, and a curated set of tools for day-to-day development automation.

## Architecture

Agentry follows the **single agent + parallel search** model. One intelligent agent handles all the work, with the ability to spawn ephemeral read-only sub-agents for parallel search when needed.

```
User Request → Main Agent (does ALL work) → Optional parallel search agents
```

## Scope

- **Runtime:** streaming loop, tool execution, budgeting, error handling
- **TUI:** conversational view with real-time streaming and reasoning display
- **LLM Providers:** OpenAI (gpt-4, gpt-5, o1, o3), Anthropic Claude, Google
- **Configuration:** single `.agentry.yaml` drives models, tools, and permissions
- **Tooling:** safe-by-default builtin tools for files, search, shell, networking
- **Observability:** JSONL traces, rolling debug logs, model pricing cache

## Vision

Deliver a dependable assistant that can plan and execute development tasks in a single repository without cloud dependencies. Keep the footprint small, make behaviour observable, and prioritise quality over breadth.

## What Exists Today

- Core conversation loop with streaming responses and structured tracing
- Multi-provider LLM support (OpenAI, Anthropic, Google)
- Reasoning model support with throttled thinking display
- Cost manager backed by a refreshable pricing cache
- Ephemeral sub-agent delegation via the `agent` tool
- Builtin TODO store with full CRUD operations
- Debug wrappers and logging infrastructure

## Recently Completed

- Simplified to single-agent architecture (removed complex multi-agent orchestration)
- Added reasoning model support with thinking content display
- Multi-provider LLM support (OpenAI, Anthropic Claude, Google)
- Removed legacy features: persistent sessions, NATS queues, Kubernetes deployment
- Comprehensive code cleanup with staticcheck analysis
- Updated documentation to reflect current architecture

## Principles

- Minimal dependencies and fast startup
- Everything observable: traces, debug logs, cost summaries
- Config-first; no hidden magic outside `.agentry.yaml`
- Documentation and tests updated alongside every feature

Keep this document terse and current—update it whenever priorities change.
