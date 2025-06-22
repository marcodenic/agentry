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

Pass `--resume-id name` to load a saved session and `--save-id name` to persist after each run.

The new `tui` command launches a split-screen interface:

```
+-------+-----------------------------+
| 🛠️ Tools | 💬 Chat / Memory           |
+-------+-----------------------------+
```

Run `agentry tui --config examples/.agentry.yaml` to try.

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
  - name: mcp # 🎮 connect to MCP servers
    type: builtin
```

The example configuration already lists these tools so they appear in the TUI's "Tools" panel. The agent decides when to use them based on model output. When no `OPENAI_KEY` is provided, the mock model only exercises the `echo` tool. To leverage the rest, set your key in `.env.local`.

Use the `mcp` tool to connect to Multi-User Connection Protocol servers. Set its
address in your YAML config and the agent can send MCP commands and read the
responses.

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

The `converse` command spawns multiple sub-agents that riff off one another. This feature is for the dev REPL only and is not available in the TUI.

```bash
converse 3 Pick a favourite movie, just one, then discuss as a group.
```

The first number selects how many agents join the chat. Any remaining text becomes the opening message. If omitted, a generic greeting is used.

### 💾 Saving & Resuming

Add a `store` path to your `.agentry.yaml` to enable persistence:

```yaml
store: path/to/db.sqlite
```

Run the CLI with `--resume-id myrun` to load a snapshot before running and `--save-id myrun` to save state after each run.

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
