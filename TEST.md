# TEST.md

**⚠️ CRITICAL: READ [CRITICAL_INSTRUCTIONS.md](./CRITICAL_INSTRUCTIONS.md) FIRST ⚠️**

> For a human-friendly guide, see [docs/testing.md](docs/testing.md).

This file is the canonical, machine-readable checklist for validating all Agentry functionality. It is optimized for agents, automation, and contributors.

---

## 1. Environment Setup

- [ ] Clone the repository
- [ ] Install Go (>=1.23)
- [ ] Install Node.js (>=18) and npm
- [ ] (Optional) Install Docker (for sandboxing, plugins)
- [ ] Copy `.env.example` to `.env.local` and set `OPENAI_KEY` if available
- [ ] Install dependencies:
  - [ ] `go mod tidy`
  - [ ] `cd ts-sdk && npm install`
  - [ ] `cd extensions/vscode-agentry && npm install`

## 2. Core Tests

- [ ] Run all Go tests: `go test ./...`
- [ ] Run all TypeScript SDK tests: `cd ts-sdk && npm test`
- [ ] Run all VS Code extension tests (if any): `cd extensions/vscode-agentry && npm test`
- [ ] Run Makefile dev target: `make dev`
- [ ] Run CLI evaluation: `agentry eval --config examples/.agentry.yaml`

## 3. End-to-End Scenarios

- [ ] Run sample flows:
  - [ ] `agentry flow examples/flows/research_task`
  - [ ] `agentry flow examples/flows/etl_pipeline`
  - [ ] `agentry flow examples/flows/multi_agent_chat`
- [ ] Run advanced demos:
  - [ ] `agentry flow agentry-demos/devops-automation`
  - [ ] `agentry flow agentry-demos/research-assistant`
- [ ] Try TUI: `agentry tui --config examples/.agentry.yaml`
- [ ] Try team chat: `agentry tui --team 3`
- [ ] Try checkpoint/resume: `agentry flow ... --checkpoint-id test`

## 4. Built-in Tools & Plugins

- [ ] Validate all built-in tools (echo, ping, bash, fetch, glob, grep, ls, view, write, edit, patch, sourcegraph, agent, mcp)
- [ ] Test plugin loading from registry: `agentry plugin fetch examples/registry/index.json example`
- [ ] Test OpenAPI/MCP tool generation: see `examples/echo-openapi.yaml`, `examples/ping-mcp.json`

## 5. Memory & Vector Store

- [ ] Test memory backends: SQLite, file, in-memory
- [ ] Test vector store integrations: qdrant, faiss, in-memory

## 6. SDK & Extension

- [ ] Run SDK example: `node -e "const {invoke}=require('./dist/index.js');invoke('hi',{stream:false}).then(console.log)"` (from `ts-sdk`)
- [ ] Build and run VS Code extension, connect to running server

## 7. Distributed/Server Testing (Optional)

- [ ] Install on Linux server (see README)
- [ ] Run flows and scenarios at scale
- [ ] Validate performance, memory, and vector store integrations

## 8. CI/CD

- [ ] Ensure all tests pass in GitHub Actions
- [ ] Review test coverage and add missing tests

---

> For troubleshooting, environment details, and human-friendly explanations, see [docs/testing.md](docs/testing.md).
