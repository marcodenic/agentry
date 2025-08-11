# Usage

## Quick Start

```bash
go install github.com/marcodenic/agentry/cmd/agentry@latest
```

The `examples/.agentry.yaml` file contains a ready-to-use configuration for these commands.

Core subcommands:


Use `--port 9090` or set `AGENTRY_PORT` to change the HTTP server port. Set
Agents run until they produce a final answer; there is no built-in iteration cap.
 
Examples:

```bash
# one-shot, run Agent 0
agentry invoke "summarize README"

# delegate directly to a role by name
agentry invoke --agent coder "add a Makefile target"

# trace to a JSONL file
agentry invoke --trace trace.jsonl "explain the code"

# team ops
agentry team roles
agentry team spawn --name coder --role coder
agentry team call --agent coder --input "print hello world"

# memory export/import (stub)
agentry memory export --out mem.json
agentry memory import --in mem.json
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

Each pane includes a real-time dashboard showing a token usage bar and a sparkline of recent activity for that agent.

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
  "idleColor": "#22C55E",
  "runningColor": "#FBBF24",
  "errorColor": "#EF4444",
  "stoppedColor": "#6B7280",
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

Individual tools can include their own permissions block. Setting `allow: false` disables that tool:

```yaml
tools:
  - name: echo
    type: builtin
    permissions:
      allow: false
```

Set `AGENTRY_CONFIRM=1` to require confirmation before overwriting files. Tool executions can be logged by setting `AGENTRY_AUDIT_LOG=path/to/audit.jsonl`.

## Observability

Enable Prometheus metrics and OTLP traces in your config:

```yaml
metrics: true
collector: localhost:4318
```

You can override the collector address via the `AGENTRY_COLLECTOR` environment
variable:

```bash
export AGENTRY_COLLECTOR=collector.example.com:4318
```

The server then exposes `/metrics` and streams spans to the specified collector.

Metrics include HTTP request counts (`agentry_http_requests_total`),
token usage (`agentry_tokens_total`) and tool execution latency
(`agentry_tool_latency_seconds`).

### Cost Analysis

Use `agentry cost` to summarize token usage and estimated cost from a
JSONL trace log:

```bash
agentry cost --input "original prompt" trace.jsonl
```

The command prints the total tokens processed and approximate dollar cost.

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

## Trace Log Analysis

If tracing is enabled via `AGENTRY_TRACE_FILE`, analyze the resulting log after a run:

```bash
agentry analyze path/to/trace.jsonl
```

## Profiling

Use `agentry pprof` to explore profiling data in your browser:

```bash
agentry pprof cpu.out
```

The command launches `go tool pprof -http` on the given profile file and blocks until you exit.

## Agent Lifecycle: Persistent vs. Session-Based

A fundamental concept in `agentry` is the lifecycle of an agent. You can run agents in two primary modes:

### 1. Session-Based Mode (Default)

When you use a configuration like `smart-config.yaml`, agents are created for a single session. They have **conversation memory** for the duration of the task, allowing them to handle follow-up instructions and maintain context. Once the task is complete, the agents and their memory are discarded.

- **Use Case:** Ideal for one-off tasks, development, and testing. It's like hiring a consultant for a specific project.

### 2. Persistent Mode

Enabled by `persistent-config.yaml`, this mode transforms agents into **long-running, stateful services**. They are not discarded after a task. Instead, they maintain their state and memory indefinitely, listening on network ports for new instructions.

- **Use Case:** Essential for building systems where agents need to be "always-on." For example, an agent that continuously monitors a system, manages a long-term project, or needs to be available for asynchronous communication from other services. It's like having a full-time employee who is always available.
