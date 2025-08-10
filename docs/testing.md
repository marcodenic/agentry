# Testing & Validation Guide

### TypeScript SDK

```bash
cd ts-sdk
npm test
```

For the canonical, up-to-date test checklist, see [../TEST.md](../TEST.md).

This guide explains how to set up your environment, run all tests, and validate Agentry functionality. It is intended for human users and new contributors.

---

## Quick Start

1. **Clone the repository**
2. **Install prerequisites:**
   - Go (>=1.23)
   - Node.js (>=18) and npm
   - (Optional) Docker (for sandboxing, plugins)
3. **Copy `.env.example` to `.env.local`** and set your `OPENAI_API_KEY` if you have one.
4. **Install dependencies:**
   - `go mod tidy`
   - `cd ts-sdk && npm install`

---

## Running Tests

### Go Tests

```bash
go test ./...
```

### TypeScript SDK Tests

```bash
cd ts-sdk
npm test
```

### VS Code Extension

```bash
cd extensions/vscode-agentry
npm test
```

### All-in-One (Makefile)

```bash
make dev
```

---

## End-to-End Scenarios

- Try the TUI: `agentry tui --config examples/.agentry.yaml`
- Team ops: `agentry team roles`, `agentry team spawn --name coder --role coder`, `agentry team call --agent coder --input "hi"`
- One-shot: `agentry invoke "say hi"`, `agentry invoke --agent coder "write hello.go"`

---

## Built-in Tools & Plugins

- Validate all built-in tools (see README for full list)
- Test OpenAPI/MCP tool generation: see `examples/echo-openapi.yaml`, `examples/ping-mcp.json`

---

## Memory & Vector Store

- Test memory backends: SQLite, file, in-memory
- Test vector store integrations: qdrant, faiss, in-memory

---

## SDK & VS Code Extension

- Run SDK example:
  ```bash
  cd ts-sdk
  node -e "const {invoke}=require('./dist/index.js');invoke('hi',{stream:false}).then(console.log)"
  ```
- Build and run VS Code extension, connect to running server

---

## Distributed/Server Testing

- Install on Linux server (see README)
- Run flows and scenarios at scale
- Validate performance, memory, and vector store integrations

---

## CI/CD

- All tests are run in GitHub Actions
- Review test coverage and add missing tests as needed

---

## Troubleshooting

- If you encounter errors, check your environment variables and dependencies
- For Windows users, ensure PowerShell is available
- For Docker-related issues, verify Docker is running and accessible
- For more help, see the README or open an issue

---

## 🪟 Windows Setup & NATS Server

Some tests and features require a running NATS server. On Windows:

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
