# Usage

## Quick Start

```bash
# CLI dev REPL with tracing
go install github.com/marcodenic/agentry/cmd/agentry@latest
agentry dev

# HTTP server + JS client
agentry serve --config examples/.agentry.yaml
npm i @marcodenic/agentry
```

The `examples/.agentry.yaml` file contains a ready-to-use configuration for these commands.

You can now use subcommands instead of the --mode flag:

- `agentry dev` (REPL)
- `agentry serve` (HTTP server)
- `agentry tui` (TUI interface)
- `agentry eval` (evaluation)
- `agentry flow` (run `.agentry.flow.yaml`)

Example:

```bash
agentry flow .
```

Run the sample scenarios in `examples/flows`:

```bash
agentry flow examples/flows/research_task
agentry flow examples/flows/etl_pipeline
agentry flow examples/flows/multi_agent_chat
```

Pass `--resume-id name` to load a saved session and `--save-id name` to persist after each run.
Use `--checkpoint-id name` to continuously snapshot the run loop and resume after a crash.

### Terminal UI

Start the interactive interface:

```bash
agentry tui --config examples/.agentry.yaml
```

There is no `--team` flag. From inside the chat you can spawn additional agents at any time:

```bash
/spawn coder "handle all build tasks"
```

Spawned agents appear in their own panes and may run on remote nodes if your Agentry cluster is configured.

#### TUI Commands

Inside the chat input you can control running agents:

- `/spawn <name>` – create another agent pane
- `/switch <prefix>` – focus an agent by ID prefix
- `/stop <prefix>` – halt an agent and keep its history
- `/converse <n> <topic>` – open a side conversation between `n` new agents

### TUI Themes & Keybinds

Create a `theme.json` file to customise colours and keyboard shortcuts. Agentry looks for the file in the current directory and its parents, falling back to `$HOME/.config/agentry/theme.json`.

```json
{
  "userBarColor": "#00FF00",
  "aiBarColor": "#FF00FF",
  "keybinds": {
    "quit": "ctrl+c",
    "toggleTab": "tab",
    "submit": "enter"
  }
}
```

## Agent Delegation

Planners can offload work to another agent using the `agent` tool. Add it to
your configuration:

```yaml
tools:
  - name: agent
    type: builtin
```

Invoke the tool with the target agent and task:

```bash
agent --agent coder --task "draft documentation"
```

## Git Branch Management

The `branch-tidy` tool helps clean up local Git repositories by removing old branches:

```yaml
tools:
  - name: branch-tidy
    type: builtin
```

The tool provides several options:

- `dry-run`: Preview which branches would be deleted without actually deleting them
- `force`: Use Git's `-D` flag instead of `-d` for force deletion

The tool automatically protects common branches (`main`, `master`, `develop`, `development`) and the current working branch.

Example usage:

```bash
# Preview what would be deleted
branch-tidy --dry-run true

# Delete branches with confirmation (safe delete)
branch-tidy --force false

# Force delete all eligible branches
branch-tidy --force true
```

## Security

Define a `permissions` section in `.agentry.yaml` to restrict which builtin tools may run:

```yaml
permissions:
  tools:
    - echo
    - ls
```

Set `AGENTRY_CONFIRM=1` to require confirmation before overwriting files. Tool executions can be logged by setting `AGENTRY_AUDIT_LOG=path/to/audit.jsonl`.

## Observability

Enable Prometheus metrics and OTLP traces in your config:

```yaml
metrics: true
collector: localhost:4318
```

The server then exposes `/metrics` and streams spans to the specified collector.

## Plugin Management

Agentry includes tooling to fetch and install external plugins:

```bash
agentry plugin fetch docs/registry/plugins.json agentry-shell
agentry plugin install https://github.com/marcodenic/agentry-shell
```

Set `AGENTRY_REGISTRY_GPG_KEYRING` to the exported public key to enable
signature verification:

```bash
export AGENTRY_REGISTRY_GPG_KEYRING=docs/registry/registry.pub
```

Create new tools with `agentry tool init <name>`. Downloaded plugins are verified against the registry's signature before installation.
