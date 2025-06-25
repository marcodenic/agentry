# AGENTS.md

> **Looking for the high-level roadmap, strategic context, or upcoming features?**
> See [ROADMAP.md](./ROADMAP.md) in the project root. ROADMAP.md contains the master plan, forward-looking goals, and detailed roadmap for Agentry Cloud—including persistent memory, workflow orchestration, sandboxing, distributed scheduling, and more.

---

## Repository Organization & Best Practices

### 📁 **Directory Structure Guidelines**

Follow Go project best practices and keep the repository organized:

#### **Core Directories**

- `cmd/` - Main applications and entry points
- `internal/` - Private application and library code
- `pkg/` - Library code that's ok to use by external applications
- `api/` - Protocol definition files (e.g., OpenAPI/Swagger specs, protocol buffers)

#### **Documentation & Examples**

- `docs/` - Documentation files
- `examples/` - Example configurations and use cases
- `templates/` - Template files for agents, teams, and roles

#### **Testing & Development**

- `tests/` - Integration tests and test utilities
- `test-programs/` - Test programs and scenarios (organized by purpose)
- `scripts/` - Build and deployment scripts
- `.github/` - GitHub workflows and templates

#### **Deployment & Packaging**

- `deploy/` - Deployment configurations (k8s, helm, docker)
- `packaging/` - Package creation scripts (homebrew, scoop, deb)

#### **Development Tools**

- `extensions/` - IDE extensions and tools
- `ui/` - Web interface and frontend code
- `ts-sdk/` - TypeScript SDK

### 🚫 **What NOT to put in the root directory:**

- ❌ Test files (`test_*.go`, `*_test.yaml`)
- ❌ Debug scripts (`debug_*.go`, `test_*.ps1`)
- ❌ Temporary configuration files
- ❌ Verification or experiment files

### ✅ **Proper places for test files:**

- **Unit tests**: Place `*_test.go` files next to the code they test
- **Integration tests**: Use `tests/` directory with descriptive subdirectories
- **Test configurations**: Use `tests/configs/` or `examples/`
- **Test programs**: Use `test-programs/` with clear naming

### 📝 **File Naming Conventions**

- Use kebab-case for YAML/config files: `agent-config.yaml`
- Use snake_case for Go files: `agent_test.go`
- Use descriptive names that indicate purpose: `integration_test.go`, `benchmark_test.go`
- Avoid generic names like `test.go`, `debug.go`, `temp.yaml`

### 🧹 **Repository Cleanup Guidelines**

- **Test files**: Always place test files in appropriate directories (`tests/`, `test-programs/`)
- **Configurations**: Store test configurations in `tests/configs/` or `examples/`
- **Debug artifacts**: Remove debug files immediately after use
- **Temporary files**: Never commit temporary or experimental files to the repository

### 🛠 **Cross-Platform Development Notes**

- **Windows PowerShell**: Use full path `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe` in Go code
- **Command translation**: Unix commands are automatically translated to PowerShell on Windows
- **Path separators**: Use `filepath.Join()` for cross-platform path handling
- **Shell execution**: Prefer `sbox.ExecDirect()` with proper OS detection

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

| Name           | Description               | Entrypoint                  | Example Usage                 |
| -------------- | ------------------------- | --------------------------- | ----------------------------- |
| agentry        | Main agent runtime        | cmd/agentry/main.go         | `go run cmd/agentry/main.go`  |
| ts-sdk         | TypeScript SDK entrypoint | ts-sdk/src/index.ts         | `cd ts-sdk && npm run build`  |
| trace analyzer | Token usage summary tool  | cmd/agentry/main.go analyze | `agentry analyze trace.jsonl` |
| pprof viewer   | Profile inspection        | cmd/agentry/main.go pprof   | `agentry pprof cpu.out`       |

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
