# Usage

## Quick Start

```bash
go install github.com/marcodenic/agentry/cmd/agentry@latest
```

A ready-to-use `.agentry.yaml` ships in the repository root. Copy it (or create a project-specific variant) and list any additional tools/roles you need.

## Running Agentry

Agentry defaults to the terminal UI when no subcommand is provided:

```bash
agentry                 # start the TUI
agentry "summarize README"  # run a one-shot prompt
agentry refresh-models      # refresh cached pricing data
```

Useful flags:

- `--config PATH` – choose a different `.agentry.yaml`
- `--debug` – enable verbose stdout logging
- `--trace trace.jsonl` – write structured events to a JSONL file
- `--allow-tools/--deny-tools/--disable-tools` – override tool permissions at runtime
- `--max-iter N` – cap iterations (default: unlimited unless budget exceeded)

## Terminal UI Highlights

- Chat pane for Agent 0
- Tool output pane showing recent invocations
- Status footer with token usage and elapsed time
- TODO board fed by the built-in TODO store

Use `tab` to switch panes. Type `/help` inside the chat for the latest keybinds and slash commands.

## Configuring Tools & Permissions

Tools are enabled by listing them in `.agentry.yaml`. Example:

```yaml
tools:
  - name: view
    type: builtin
  - name: create
    type: builtin
  - name: agent
    type: builtin

permissions:
  tools:
    - view
    - create
    - agent
```

You can additionally gate tools per-entry:

```yaml
tools:
  - name: bash
    type: builtin
    permissions:
      allow: false
```

At runtime, combine the config with CLI flags for final allow/deny behaviour.

## Tracing & Debugging

- Pass `--trace trace.jsonl` to capture structured events for later analysis
- Set `AGENTRY_DEBUG_LEVEL=trace` for verbose output (works with the TUI)
- `scripts/debug-agentry.sh` wraps `agentry` with trace-friendly defaults
- `scripts/test_debug_logging.sh` runs a quick smoke test against the logging stack

## Model Pricing Cache

`agentry refresh-models` downloads current pricing information (via models.dev) and stores it in the user cache directory. The runtime uses this cache when calculating usage costs during a session.

## Environment Variables

Copy `.env.example` to `.env.local` and set:

- `OPENAI_API_KEY` – required for OpenAI-backed models
- `ANTHROPIC_API_KEY` (optional) – enable Anthropic models when configured

`.env.local` is loaded automatically at startup.

## Observability & Costs

Every run prints a summary of input/output tokens and the estimated dollar amount. Combine trace logs with the pricing cache to audit multi-step workflows.

## Related Documents

- [Configuration Guide](CONFIG_GUIDE.md)
- [API / tool catalogue](api.md)
- [Debug logging](DEBUG_LOGGING.md)
- [Testing guide](testing.md)
