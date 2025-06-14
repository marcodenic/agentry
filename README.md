# Agentry\u00a0\u2013 Minimal, Performant AI-Agent Framework (Go core + TS SDK)

Agentry is a production-ready **agent runtime** written in Go with an optional TypeScript client.

| Pillar            | v\u00a01.0 Features                                      |
| ----------------- | -------------------------------------------------------- |
| **Minimal core**  | ~200\u00a0LOC run loop, zero heavy deps                  |
| **Plugins**       | JSON/YAML tool manifests; Go or external processes       |
| **Sub-agents**    | `Spawn()` + `RunParallel()` helper                       |
| **Model routing** | Rule-based selector, multi-LLM support                   |
| **Memory**        | Conversation + VectorStore interface (RAG-ready)         |
| **Tracing**       | Structured events, JSONL dump, SSE stream                |
| **Config**        | `.agentry.yaml` bootstraps agent, models, tools          |
| **Evaluation**    | YAML test suites, CLI `agentry eval`                     |
| **SDK**           | JS/TS client (`@marcodenic/agentry`), supports streaming |

See docs/ for full guides. Quick start:

```bash
# CLI dev REPL with tracing
go install github.com/marcodenic/agentry/cmd/agentry@latest
agentry dev

# HTTP server + JS client
agentry serve --config .agentry.yaml
npm i @marcodenic/agentry
```

### Try it live

```bash
# one-off REPL (OpenAI key picked up from .env.local)
agentry dev               # type messages, see responses

# HTTP + TS SDK
agentry serve --config examples/.agentry.yaml &
npm --prefix ts-sdk install
npm --prefix ts-sdk run build
node -e "const {invoke}=require('./ts-sdk/dist');invoke('hi',{stream:false}).then(console.log)"
```

## Environment Configuration

Copy `.env.example` to `.env.local` and fill in `OPENAI_KEY` to enable real OpenAI calls. The file is loaded automatically on startup and during tests.

To run evaluation with the real model:

```bash
OPENAI_KEY=your-key agentry --mode=eval --config my.agentry.yaml
```

When the real model is active, the CLI uses `tests/openai_eval_suite.json` so the
assertions match ChatGPT's typical response.

Evaluation results are printed to the console when using this mode.

If no key is present, the built-in mock model is used.

## Testing & Development

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
