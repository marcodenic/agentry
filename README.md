# 🤖 Agentry — Minimal, Performant AI-Agent Framework (Go core + TS SDK)

Agentry is a production-ready **agent runtime** written in Go with an optional TypeScript client.

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

See `docs/` for full guides.

```bash
# 🖥️ CLI dev REPL with tracing
go install github.com/marcodenic/agentry/cmd/agentry@latest
agentry dev

# 🌐 HTTP server + JS client
agentry serve --config .agentry.yaml
npm i @marcodenic/agentry
```

You can now use subcommands instead of the --mode flag:

- `agentry dev` (REPL)
- `agentry serve` (HTTP server)
- `agentry tui` (TUI interface)
- `agentry eval` (evaluation)

The new `tui` command launches a split-screen interface:

```
+-------+-----------------------------+
| 🛠️ Tools | 💬 Chat / Memory           |
+-------+-----------------------------+
```

Run `agentry tui --config examples/.agentry.yaml` to try.

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
```

The example configuration already lists these tools so they appear in the TUI's "Tools" panel. The agent decides when to use them based on model output. When no `OPENAI_KEY` is provided, the mock model only exercises the `echo` tool. To leverage the rest, set your key in `.env.local`.

> **🪟 Windows users:** Agentry works out-of-the-box on Windows 10+ with PowerShell installed. Built-ins that require external Unix tools (`patch`) are disabled automatically. Install Git for Windows and run under Git Bash if you need them.

---

## 🧑‍💻 Try it Live

```bash
# 🏃‍♂️ One-off REPL (OpenAI key picked up from .env.local)
agentry dev               # type messages, see responses

# 🌐 HTTP + TS SDK
agentry serve --config examples/.agentry.yaml &
npm --prefix ts-sdk install
npm --prefix ts-sdk run build
node -e "const {invoke}=require('./ts-sdk/dist');invoke('hi',{stream:false}).then(console.log)"
```

---

## 🛠️ Dev REPL Tricks

### 🤖 Multi-agent conversations

The `converse` command spawns multiple sub-agents that riff off one another. This feature is for the dev REPL only and is not available in the TUI.

```bash
converse 3 Pick a favourite movie, just one, then discuss as a group.
```

The first number selects how many agents join the chat. Any remaining text becomes the opening message. If omitted, a generic greeting is used.

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
