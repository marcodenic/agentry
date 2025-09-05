Agentry — Minimal, extensible AI agent runtime for the terminal

Overview
- Minimal core in Go with fast startup and no heavy deps
- Built-in TUI for day-to-day coding and debugging
- Pluggable tools via simple manifests; safe permission gating
- Team/delegation helpers to spawn and coordinate sub-agents
- Model-agnostic with mock mode for offline/dev
- Structured tracing and live token/cost accounting

Demo: see `agentry.gif` in the repo

Install
- Prereq: Go 1.23+
- Install CLI: `go install github.com/marcodenic/agentry/cmd/agentry@latest`

Quick Start
- TUI (recommended): `agentry tui --config examples/.agentry.yaml`
- Direct prompt: `agentry "summarize the README"`
- Show version: `agentry --version`

Configuration
- Project config: `.agentry.yaml` (an example lives in `examples/.agentry.yaml`)
- Env vars: copy `.env.example` to `.env.local` and set keys (e.g., `OPENAI_API_KEY`)
- Flags you may care about:
  - `--config PATH`: select config file
  - `--theme THEME`: theme override for TUI
  - `--debug`: verbose diagnostics
  - `--allow-tools a,b` / `--deny-tools a,b` / `--disable-tools`
  - `--resume-id` / `--save-id` / `--checkpoint-id` for session state
  - `--port` to set embedded HTTP server

Usage Notes
- TUI launches when no command is provided: just run `agentry`
- You can also pass a direct prompt without a subcommand
- The TUI supports spawning additional agents and shows live token/cost usage

Built-in Tools
- Tools are enabled by listing them in your `.agentry.yaml` and permissions
- Common tools include: `echo`, `ping`, `view`, `write`, `edit`, `patch`, `grep`, `ls`, `agent`, `mcp`
- Permissions allow you to strictly gate what the agent can use

Tracing & Costs
- Every run can emit structured trace events
- Summaries include input/output tokens and estimated cost per run

Development
- Build: `make build` (outputs `./agentry`)
- Tests: `go test ./...` or `./scripts/test.sh`
- Formatting: CI enforces `gofmt -l` cleanliness

Versioning & Releases
- The internal version constant lives in `internal/version.go`
- Release workflow publishes binaries on tag push like `v0.1.1`

License
- MIT, see `LICENSE`
