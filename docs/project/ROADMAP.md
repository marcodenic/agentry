# Roadmap

_Last updated: December 2025_

## Now (in flight)

- Polish TUI experience and navigation
- Expand test coverage for tool execution paths
- Performance tuning for reasoning model streaming

## Next (upcoming)

- Additional LLM provider integrations
- Enhanced TODO board with filtering
- Workflow presets for common repo setups (Go, JS/TS, Python)

## Later (nice to have)

- Optional telemetry export building on JSONL trace pipeline
- Library mode for embedding the runtime in other Go projects
- MCP server implementations for tool extensibility

## Completed

- Simplified to single-agent + parallel search architecture
- Multi-provider LLM support (OpenAI, Anthropic, Google)
- Reasoning model support with throttled thinking display
- Removed legacy multi-agent orchestration complexity
- Comprehensive code cleanup with staticcheck
- Updated all documentation to reflect current architecture

## Principles

- Stay local-first, minimal-dependency
- Document everything; configuration lives in `.agentry.yaml`
- Ship traceable, testable features before adding new surface area

Updates to this file should reflect meaningful shifts in focus or newly delivered work.
