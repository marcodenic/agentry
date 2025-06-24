# 🤖 Agentry — Minimal, Performant AI-Agent Framework (Go core + TS SDK)

![Demo](agentry.gif)

Agentry is a production-ready **agent runtime** written in Go with an optional TypeScript client.

For the upcoming cloud deployment model, see [README-cloud.md](./README-cloud.md).

---

| 🚩 **Pillar**        | ✨ **v1.0 Features**                                     |
| -------------------- | -------------------------------------------------------- |
| 🦴 **Minimal core**  | ~200 LOC run loop, zero heavy deps                       |
| 🔌 **Plugins**       | JSON/YAML tool manifests; Go or external processes       |
| 🤹‍♂️ **Sub-agents**    | `Spawn()` + `RunParallel()` helper                       |
| 🧭 **Model routing** | Rule-based selector, multi-LLM support                   |
| 🧠 **Memory**        | Conversation + VectorStore interface (RAG-ready)         |
| 🕵️‍♂️ **Tracing**       | Structured events, JSONL dump, SSE stream                |
| ⚙️ **Config**        | `.agentry.yaml` bootstraps agent, models, tools          |
| 🧪 **Evaluation**    | YAML test suites, CLI `agentry eval`                     |
| 🛠️ **SDK**           | JS/TS client (`@marcodenic/agentry`), supports streaming |
| 📦 **Registry**      | [Plugin Registry](docs/registry/)                        |

---

## 📦 Installation

Prebuilt binaries are available on [GitHub Releases](https://github.com/marcodenic/agentry/releases).

### Homebrew (macOS/Linux)

```bash
brew tap marcodenic/agentry
brew install agentry
```

### Scoop (Windows)

```powershell
scoop bucket add agentry https://github.com/marcodenic/agentry
scoop install agentry
```

### Debian

```bash
wget https://github.com/marcodenic/agentry/releases/download/vX.Y.Z/agentry_X.Y.Z_amd64.deb
sudo dpkg -i agentry_X.Y.Z_amd64.deb
```

---

## 🚀 Quick Start

```bash
# 🖥️ CLI dev REPL with tracing
go install github.com/marcodenic/agentry/cmd/agentry@latest
agentry dev

# 🌐 HTTP server + JS client
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

More advanced scenarios are available in the [agentry-demos](./agentry-demos) repository:

```bash
agentry flow agentry-demos/devops-automation
agentry flow agentry-demos/research-assistant
```

Pass `--resume-id name` to load a saved session and `--save-id name` to persist after each run.
Use `--checkpoint-id name` to continuously snapshot the run loop and resume after a crash.

The new `tui` command launches a split-screen interface:

```
+-------+-----------------------------+
| 🛠️ Tools | 💬 Chat / Memory           |
+-------+-----------------------------+
```

Run `agentry tui --config examples/.agentry.yaml` to try. Use `--team 3` to launch team chat mode where each agent has its own pane.

---

### 🎨 Themes & Keybinds

Create a `theme.json` file to customise colours and keyboard shortcuts. Agentry
looks for the file in the current directory and its parents, falling back to
`$HOME/.config/agentry/theme.json`. Settings override the built‑in defaults.

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

---

## 🧰 Built-in Tools

Agentry ships with a collection of safe builtin tools. They become available to the agent when listed in your `.agentry.yaml` file:

```yaml
tools:
  - name: echo # 🔁 repeat a string
    type: builtin
  - name: ping # 📡 ping a host
    type: builtin
  - name: bash # 🖥️ run a bash command
    type: builtin
  - name: fetch # 🌍 download content from a URL
    type: builtin
  - name: glob # 🗂️ find files by pattern
    type: builtin
  - name: grep # 🔎 search file contents
    type: builtin
  - name: ls # 📁 list directory contents
    type: builtin
  - name: view # 👀 read a file
    type: builtin
  - name: write # ✍️ create or overwrite a file
    type: builtin
  - name: edit # 📝 update an existing file
    type: builtin
  - name: patch # 🩹 apply a unified diff
    type: builtin
  - name: sourcegraph # 🔍 search public repositories
    type: builtin
  - name: agent # 🤖 launch a search agent
    type: builtin
  - name: delegate # 📋 assign tasks to another agent
    type: builtin
  - name: mcp # 🎮 connect to MCP servers
    type: builtin
```

The example configuration already lists these tools so they appear in the TUI's "Tools" panel. The agent decides when to use them based on model output. When no `OPENAI_KEY` is provided, the mock model only exercises the `echo` tool. To leverage the rest, set your key in `.env.local`.

Use the `mcp` tool to connect to Multi-User Connection Protocol servers. Set its
address in your YAML config and the agent can send MCP commands and read the
responses.

### 📋 Delegating Tasks

Planners can forward work to specialised agents using the new `delegate` tool.
Include it in your `.agentry.yaml` and call it with the target agent name and
task description:

```bash
delegate --agent coder --task "write unit tests"
```

### OpenAPI & MCP Specs

Agentry can generate tools from an OpenAPI document or a simple MCP schema. Use
`tool.FromOpenAPI` or `tool.FromMCP` to load a spec and obtain a registry of
HTTP-backed tools. Example specs are provided in `examples/echo-openapi.yaml` and
`examples/ping-mcp.json`.

> **🪟 Windows users:** Agentry works out-of-the-box on Windows 10+ with PowerShell installed.

---

## 🧑‍💻 Try it Live

```bash
# 🏃‍♂️ One-off REPL (OpenAI key picked up from .env.local)
agentry dev               # type messages, see responses

# 🌐 HTTP + TS SDK
agentry serve --config examples/.agentry.yaml &

# In a new terminal, run the following from the ts-sdk directory:
cd ts-sdk
npm install  # (if you get dependency errors, use: npm install --legacy-peer-deps)
npm run build
npm test

# Make sure the agentry server is running, then:
node -e "const {invoke}=require('./dist/index.js');invoke('hi',{stream:false}).then(console.log)"
```

---

## 🦾 Full End-to-End Example (Two Terminals)

> **You must use two terminals for this demo.**
>
> - **Terminal 1:** Start the Agentry server from the project root.
> - **Terminal 2:** Run the TypeScript SDK example from the `ts-sdk` directory.

### 🖥️ Terminal 1: Start the Agentry server

**From the project root directory (`agentry`):**

```bash
agentry serve --config examples/.agentry.yaml
```

- This will start the server and take over the terminal. Leave it running.

### 🖥️ Terminal 2: Use the TypeScript SDK

**From the `ts-sdk` directory:**

If you are not already in the `ts-sdk` directory, run:

```bash
cd ts-sdk
```

Then, from inside `ts-sdk`, run each command one at a time:

```bash
npm install  # (if you get dependency errors, use: npm install --legacy-peer-deps)
npm run build
npm test
node -e "const {invoke}=require('./dist/index.js');invoke('hi',{stream:false,agentId:'default'}).then(console.log).catch(e=>console.error(e.message))"
```

- **Do not run `cd ts-sdk` if you are already in the `ts-sdk` directory.**
- All npm commands and the Node.js example must be run from inside `ts-sdk`.
- Make sure the server in Terminal 1 is running before running the Node.js command above.
- If you see a connection error, check that Terminal 1 is still running and listening on port 8080.

---

## 🛠️ Dev REPL Tricks

### 🤖 Multi-agent conversations

The `converse` command spawns multiple sub-agents that riff off one another. This was originally REPL-only, but the TUI now supports team chat via `--team`.

```bash
converse 3 Pick a favourite movie, just one, then discuss as a group.
```

The first number selects how many agents join the chat. Any remaining text becomes the opening message. If omitted, a generic greeting is used.

### 💾 Saving & Resuming

Add a `memory` entry to your `.agentry.yaml` to enable persistence. The value uses a URI scheme to select the backend:

```yaml
# SQLite database
memory: sqlite:mem.db

# JSON file
# memory: file:mem.json

# In-memory (ephemeral)
# memory: mem:

# Session TTL (optional)
store: path/to/db.sqlite
# automatically remove sessions after one week
session_ttl: 168h
```

Run the CLI with `--resume-id myrun` to load a snapshot before running and `--save-id myrun` to save state after each run. `--checkpoint-id myrun` continuously saves intermediate steps so sessions can be resumed.
Expired sessions are pruned automatically by the server based on `session_ttl`.
Cleanup runs hourly for any configured store backend.

### 📚 Vector Store

Configure a vector backend for document retrieval:

```yaml
vector_store:
  type: qdrant
  url: http://localhost:6333
  collection: agentry
```

Supported types are `qdrant`, `faiss`, and the default in-memory store.

### ♻️ Reusing Roles

Role templates live under `templates/roles/`. Each YAML file defines an agent
name, prompt, and allowed tools:

```yaml
name: coder
prompt: |
  You are an expert software developer.
tools:
  - bash
  - patch
```

Reference templates from a flow using the `include` key:

```yaml
include:
  - templates/roles/coder.yaml

agents:
  coder:
    model: gpt-4o

tasks:
  - agent: coder
    input: build a CLI
```

The template's prompt and tools merge with the agent definition. Paths are
resolved relative to the flow file.

---

## ⚙️ Environment Configuration

Copy `.env.example` to `.env.local` and fill in `OPENAI_KEY` to enable real OpenAI calls. The file is loaded automatically on startup and during tests.

To run evaluation with the real model:

```bash
OPENAI_KEY=your-key agentry eval --config my.agentry.yaml
```

When the real model is active, the CLI uses `tests/openai_eval_suite.json` so the assertions match ChatGPT's typical response.

Evaluation results are printed to the console when using this mode.

If no key is present, the built-in mock model is used.

---

## 🧪 Testing & Validation

- For the canonical, machine-readable checklist, see [TEST.md](./TEST.md).
- For a human-friendly guide, see [docs/testing.md](./docs/testing.md).

---

## 🧪 Testing & Development

Run all tests and start a REPL with one command:

```bash
make dev
```

This target executes Go and TypeScript tests, builds the CLI, and launches `agentry serve` using the example config. You can also run the steps manually:

```bash
go test ./...
cd ts-sdk && npm install && npm test
go install ./cmd/agentry
agentry dev
```

## 🪟 Windows Setup & NATS Server

To run tests or use Agentry features that require NATS on Windows:

1. Download and extract the latest NATS server zip from the [official releases](https://github.com/nats-io/nats-server/releases).
2. Start the server in PowerShell (adjust path as needed):
   ```powershell
   & "C:\Users\marco\Downloads\nats-server-v2.11.4-windows-amd64\nats-server-v2.11.4-windows-amd64\nats-server.exe" -p 4222
   ```

**Run Go tests (excluding integration):**

```powershell
go test ./... -v -short
```

**Run all tests, including integration:**

```powershell
go test ./... -v -tags=integration
```

---

## 🧩 VS Code Extension

The `extensions/vscode-agentry` folder contains a small helper extension that streams output from a running server.

```bash
cd extensions/vscode-agentry
npm install
npm run build
```

Start the Agentry server with `agentry serve --config examples/.agentry.yaml` then run **Agentry: Open Panel** from VS Code to connect. Use **Agentry: Stop Stream** to end the session.

## 🔌 Plugin Registry

The project publishes a small registry at `docs/registry/plugins.json`. Each
entry lists a plugin repository URL and description. Fetch a plugin and install
it with:

```bash
agentry plugin fetch docs/registry/plugins.json agentry-shell
agentry plugin install https://github.com/marcodenic/agentry-shell
```

To contribute a plugin, update `docs/registry/plugins.json` and open a pull
request.

## 🛠️ Tool Scaffolding

Create boilerplate for a new built-in tool:

```bash
agentry tool init mytool
```

This generates a folder with a Go source file and YAML manifest.

## 🔗 Wrapping OpenAPI or MCP Specs

Specs can be converted to tool definitions at the command line:

```bash
agentry tool openapi examples/echo-openapi.yaml > tools.yaml
agentry tool mcp examples/ping-mcp.json > tools.yaml
```

Each command prints YAML `ToolSpec` entries that can be inspected or embedded in
config files.
