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
| 📦 **Registry**     | [Plugin Registry](docs/registry/) |

See the [documentation](https://marcodenic.github.io/agentry/) for installation instructions, usage examples, and API details.
