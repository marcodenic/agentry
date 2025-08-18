# Agentry

Agentry is a minimal, extensible agent runtime written in Go.

| Pillar            | v1.0 Features                                            |
| ----------------- | -------------------------------------------------------- |
| **Minimal core**  | ~200 LOC run loop, zero heavy deps                       |
| **Plugins**       | JSON/YAML tool manifests; Go or external processes       |
| **Sub-agents**    | `Spawn()` + `RunParallel()` helper                       |
| **Model routing** | Rule-based selector, multi-LLM support                   |
| **Memory**        | Conversation + VectorStore interface (RAG-ready)         |
| **Tracing**       | Structured events, JSONL dump, SSE stream                |
| **Config**        | `.agentry.yaml` bootstraps agent, models, tools          |
| **Evaluation**    | YAML test suites, CLI `agentry eval`                     |
| **Costs**         | Live token & pricing accounting                          |
| **Registry**      | (Pluggable tool loading)                                 |
| **Delegation**    | `agent` tool lets planners assign tasks to agents        |

For the upcoming cloud deployment model, see [README-cloud.md](../README-cloud.md).

Check out the [installation guide](install.md) to get started.

---

## 🧪 Testing & Validation

- [TEST.md](../TEST.md): Machine-readable checklist for agents and automation.
- [Testing Guide](testing.md): Human-friendly instructions for contributors and users.
