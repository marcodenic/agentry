# Agentry\u00a0\u2013 Minimal, Performant AI-Agent Framework (Go core + TS SDK)

Agentry is a production-ready **agent runtime** written in Go with an optional TypeScript client.

| Pillar            | v\u00a01.0 Features                                     |
| ----------------- | ------------------------------------------------------- |
| **Minimal core**  | ~200\u00a0LOC run loop, zero heavy deps                 |
| **Plugins**       | JSON/YAML tool manifests; Go or external processes      |
| **Sub-agents**    | `Spawn()` + `RunParallel()` helper                      |
| **Model routing** | Rule-based selector, multi-LLM support                  |
| **Memory**        | Conversation + VectorStore interface (RAG-ready)        |
| **Tracing**       | Structured events, JSONL dump, SSE stream               |
| **Config**        | `.agentry.yaml` bootstraps agent, models, tools         |
| **Evaluation**    | YAML test suites, CLI `agentry eval`                    |
| **SDK**           | JS/TS client (`@yourScope/agentry`), supports streaming |

See docs/ for full guides. Quick start:

```bash
# CLI dev REPL with tracing
go install github.com/marcodenic/agentry/cmd/agentry@latest
agentry dev

# HTTP server + JS client
agentry serve --config .agentry.yaml
npm i @yourScope/agentry
```
