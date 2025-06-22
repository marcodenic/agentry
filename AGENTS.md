# AGENTS.md

> **Looking for the high-level roadmap, strategic context, or upcoming features?**
> See [ROADMAP.md](./ROADMAP.md) in the project root. ROADMAP.md contains the master plan, forward-looking goals, and detailed roadmap for Agentry Cloud—including persistent memory, workflow orchestration, sandboxing, distributed scheduling, and more.

---

## Project Vision

Agentry is a minimal, extensible agentic runtime written in Go, with a TypeScript SDK. The project aims to provide a robust platform for building, orchestrating, and scaling multi-agent AI systems. See [ROADMAP.md](./ROADMAP.md) for strategic direction and future plans.

---

## Agents & Entrypoints

| Name    | Description               | Entrypoint          | Example Usage                |
| ------- | ------------------------- | ------------------- | ---------------------------- |
| agentry | Main agent runtime        | cmd/agentry/main.go | `go run cmd/agentry/main.go` |
| ts-sdk  | TypeScript SDK entrypoint | ts-sdk/src/index.ts | `cd ts-sdk && npm run build` |

---

## Languages

- Go
- TypeScript

## Filetypes

- `*.go`
- `*.ts`
- `*.yaml`
- `*.json`

---

## Testing

To run all tests:

- Go: `go test ./...`
- TypeScript SDK: `cd ts-sdk && npm install && npm test`

---

## Contribution Guidelines

- To propose a new agent or entrypoint, open an issue or pull request.
- All agents/components should include tests and documentation.
- See [CONTRIBUTING.md](./CONTRIBUTING.md) for details (if available).
- Expect a response from maintainers within 7 days. If you haven’t heard back, feel free to ping the thread.

---

## Support & Communication

- For help, open an issue or join the discussion board.
- Keep communication public (issues, PRs) for transparency and shared knowledge.

---

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE).

---

## Roadmap Update Policy

- Only update the roadmap or task list in [ROADMAP.md](./ROADMAP.md) when a task has been completed to an enterprise-grade level. Partial or experimental implementations should not be marked as complete until they meet this standard.
- Review and update the roadmap as needed every time you make changes to the codebase, or when you update or create new documentation—including AI-specific documentation—to ensure alignment and accuracy.

---

_Please help keep this file up to date as agents and entrypoints are added or removed._
