# AGENTS.md

> **Looking for the high-level roadmap, strategic context, or upcoming features?**
> See [ROADMAP.md](./ROADMAP.md) in the project root. ROADMAP.md contains the master plan, forward-looking goals, and detailed roadmap for Agentry Cloud—including persistent memory, workflow orchestration, sandboxing, distributed scheduling, and more.

---

## Project Vision

Agentry is a minimal, extensible agentic runtime written in Go, with a TypeScript SDK. The project aims to provide a robust platform for building, orchestrating, and scaling multi-agent AI systems. See [ROADMAP.md](./ROADMAP.md) for strategic direction and future plans.

---

## 🚨 MANDATORY DOCUMENTATION UPDATES

**CRITICAL: When ANY work is completed, you MUST update the following files immediately:**

### ✅ REQUIRED ACTIONS AFTER COMPLETING WORK:

1. **Update [ROADMAP.md](./ROADMAP.md)**

   - Mark completed tasks as DONE
   - Remove or update obsolete items
   - Add any new tasks discovered during work
   - Update progress indicators and timelines

2. **Update Documentation**

   - Update relevant files in `/docs/` directory
   - Update API documentation if APIs changed
   - Update usage examples if functionality changed
   - Update installation/setup instructions if needed

3. **Update TODO Lists**

   - Remove completed items from any TODO lists
   - Add new discovered tasks or technical debt
   - Update priorities based on current understanding

4. **Update This File (AGENTS.md)**
   - Add new agents/entrypoints if created
   - Update examples if changed
   - Update testing instructions if new tests added

### 🎯 PURPOSE:

- **Prevent Duplicate Work**: Agents need current information to avoid repeating completed tasks
- **Maintain Accurate Backlog**: Keep the work queue current and prioritized
- **Enable Efficient Collaboration**: Ensure all agents have up-to-date context

### ⚠️ FAILURE TO UPDATE DOCUMENTATION WILL RESULT IN:

- Wasted effort on already-completed tasks
- Inconsistent project state
- Confused agents working from outdated information
- Degraded project quality and efficiency

**NO EXCEPTIONS: Documentation updates are not optional—they are a required part of completing any work.**

---

## Agents & Entrypoints

| Name    | Description               | Entrypoint          | Example Usage                |
| ------- | ------------------------- | ------------------- | ---------------------------- |
| agentry | Main agent runtime        | cmd/agentry/main.go | `go run cmd/agentry/main.go` |
| ts-sdk  | TypeScript SDK entrypoint | ts-sdk/src/index.ts | `cd ts-sdk && npm run build` |
| trace analyzer | Token usage summary tool | cmd/agentry/main.go analyze | `agentry analyze trace.jsonl` |

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

## 📋 ROADMAP & DOCUMENTATION UPDATE POLICY

**MANDATORY REQUIREMENTS:**

- **ONLY** mark tasks as complete in [ROADMAP.md](./ROADMAP.md) when they have been implemented to an **enterprise-grade level** with proper testing, documentation, and error handling
- **ALWAYS** review and update the roadmap immediately after making ANY changes to the codebase
- **ALWAYS** update relevant documentation when creating or modifying features
- **NEVER** leave partial or experimental implementations marked as complete
- **VERIFY** that all related documentation accurately reflects the current state after your changes

**This is not a suggestion—it is a requirement for all contributors and agents.**

---

_Please help keep this file up to date as agents and entrypoints are added or removed._
