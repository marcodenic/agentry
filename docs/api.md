# API Overview

Agentry ships with a curated set of builtin tools. Tools are only available when listed in `.agentry.yaml` and permitted by the `permissions` block or runtime flags.

```yaml
tools:
  - name: view
    type: builtin
  - name: create
    type: builtin
  - name: edit_range
    type: builtin
  - name: search_replace
    type: builtin
  - name: ls
    type: builtin
  - name: find
    type: builtin
  - name: grep
    type: builtin
  - name: glob
    type: builtin
  - name: fetch
    type: builtin
  - name: api
    type: builtin
  - name: download
    type: builtin
  - name: bash
    type: builtin
  - name: sh
    type: builtin
  - name: powershell
    type: builtin
  - name: cmd
    type: builtin
  - name: lsp_diagnostics
    type: builtin
  - name: agent
    type: builtin
```

### Tool Categories

| Category         | Tools (examples)                              |
| ---------------- | ---------------------------------------------- |
| File operations  | `view`, `create`, `edit_range`, `search_replace` |
| Search           | `ls`, `find`, `grep`, `glob`                   |
| Networking       | `fetch`, `api`, `download`                     |
| Shell            | `bash`, `sh`, `powershell`, `cmd`              |
| Diagnostics      | `lsp_diagnostics`, `sysinfo`, `ping`           |
| Delegation       | `agent`                                        |

Include only the tools you trust for a repository. Use runtime flags (`--allow-tools`, `--deny-tools`, `--disable-tools`) for temporary overrides during a session.

### Environment

Copy `.env.example` to `.env.local` and set relevant keys:

- `OPENAI_API_KEY` for OpenAI-backed models (required for non-mock runs)
- `ANTHROPIC_API_KEY` if you opt into Anthropic models

`.env.local` is loaded automatically at startup.

## CLI Surface

The CLI is intentionally small:

| Command                  | Description                              |
| ------------------------ | ---------------------------------------- |
| `agentry`                | Launch the TUI (default)                 |
| `agentry "prompt"`       | Run a direct one-shot prompt             |
| `agentry refresh-models` | Refresh cached pricing data              |
| `agentry --help`         | Print usage information                  |

All output is JSON-safe (no extra shell colour codes). Combine commands with `--trace` when you need structured logs.

## Tracing & Costs

- `--trace trace.jsonl` writes streaming events compatible with JSONL tooling
- The runtime prints a token/cost summary after each invocation
- Pricing data comes from the on-disk cache populated by `refresh-models`

## Debugging Helpers

- `AGENTRY_DEBUG_LEVEL=trace` – verbose logging
- `scripts/debug-agentry.sh` – convenient wrapper with debug defaults
- `scripts/test_debug_logging.sh` – smoke test for the logging pipeline

## Extending Agentry

Builtin tools live under `internal/tool/`. Add new Go-based tools beside the existing ones or register external executables by filling out a `ToolManifest` entry in `.agentry.yaml`. Keep permissions strict and document new capabilities in the configuration guide.
