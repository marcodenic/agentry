# Roadmap

_Last updated: September 2025_

## Now (in flight)

- Finalise Context-Lite prompt compiler and associated tests
- Harden tool execution error-handling (`TreatErrorsAsResults`, budgeting)
- Audit documentation to match the trimmed CLI/TUI surface

## Next (upcoming)

- TUI improvements: TODO board filters, clearer agent status indicators
- Role/SOP refresh: standardise templates for Agent 0, Coder, Tester, Critic
- Expand integration coverage for delegation plus `AGENTRY_DELEGATION_TIMEOUT`

## Later (nice to have)

- Optional telemetry export that builds on the JSONL trace pipeline
- Library mode for embedding the runtime in other Go projects
- Workflow presets for common repo setups (Go, JS/TS, Python)

## Completed

- Removed persistent session infrastructure, NATS queueing, and autoscaler code
- Retired `examples/` configs in favour of a single root `.agentry.yaml`
- Simplified CLI to `tui`, direct prompts, and pricing refresh command
- Added regression tests for streaming aggregator and tool executor behaviour

## Principles

- Stay local-first, minimal-dependency
- Document everything; configuration lives in `.agentry.yaml`
- Ship traceable, testable features before adding new surface area

Updates to this file should reflect meaningful shifts in focus or newly delivered work.
