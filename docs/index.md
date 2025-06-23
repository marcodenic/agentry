# Agentry

Agentry is a minimal, extensible agent runtime written in Go with a TypeScript SDK.

| Pillar | v1.0 Features |
| --- | --- |
| **Minimal core** | ~200 LOC run loop, zero heavy deps |
| **Plugins** | JSON/YAML tool manifests; Go or external processes |
| **Sub-agents** | `Spawn()` + `RunParallel()` helper |
| **Model routing** | Rule-based selector, multi-LLM support |
| **Memory** | Conversation + VectorStore interface (RAG-ready) |
| **Tracing** | Structured events, JSONL dump, SSE stream |
| **Config** | `.agentry.yaml` bootstraps agent, models, tools |
| **Evaluation** | YAML test suites, CLI `agentry eval` |
| **SDK** | JS/TS client (`@marcodenic/agentry`), supports streaming |
| **Registry** | [Plugin Registry](registry/) |

For the upcoming cloud deployment model, see [README-cloud.md](../README-cloud.md).

Check out the [installation guide](install.md) to get started.
