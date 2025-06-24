# === STRATEGIC ROADMAP & CONTEXT ===

Agentry aims to become a best-in-class platform for multi-agent AI by anticipating the next wave of requirements in agentic systems. The following strategic context and recommendations inform the roadmap and epics below, ensuring Agentry can leapfrog current offerings and remain future-proof:

## Key Strategic Enhancements

1. **Persistent Memory & Resumable Workflows**

   - Enable agents to remember and resume tasks across sessions, with first-class support for persistent state storage and workflow checkpointing.
   - Integrate long-term memory (e.g., SQLite/vector DB per agent/project) and provide APIs for agents to read/write knowledge, facts, and plans.
   - Implement checkpointing and task resumption, allowing agents to serialize/deserialize state and continue complex workflows after restarts.
   - Provide example scenarios (e.g., scheduled agents, multi-day research tasks) to demonstrate reliability and persistence.

2. **Declarative Workflow Orchestration (DSL for Agents)**

   - Extend Agentry’s YAML/JSON config to support scripting of multi-agent interactions and orchestration patterns (parallel, sequential, conditional, etc.).
   - Develop a high-level, human-readable DSL for defining agent roles, communication, and tool usage, lowering the barrier for complex workflows.
   - Plan for a future no-code/low-code UI to edit and visualize workflows, making orchestration accessible to a broader audience.

3. **Enhanced Sandboxing and Plugin Ecosystem**

   - Strengthen sandboxing for tool/plugin execution (e.g., Docker, gVisor, Firecracker) to ensure system safety and configurable permissions.
   - Implement a permission system for tool usage, with interactive or policy-based approvals and comprehensive audit logging.
   - Formalize plugin APIs and align with emerging standards (OpenAI Plugin spec, Anthropics’ MCP) for maximum interoperability.
   - Foster a community-driven plugin marketplace, enabling rapid extension and sharing of tools/agents.

4. **Distributed Agent Scheduling & Collaboration**

   - Evolve Agentry into a distributed platform, orchestrating agents across processes/machines via messaging (gRPC, NATS, etc.).
   - Build a task queue and scheduler for background/long-running tasks, supporting horizontal scaling and resilience.
   - Define protocols for agent collaboration and state sharing, with optional central coordination components.
   - Integrate monitoring, discovery, and secure communication for distributed deployments.

5. **Observability and Developer Experience**
   - Provide real-time tracing, structured logging, and web dashboards for visualizing agent workflows and debugging.
   - Track metrics (token usage, latency, tool calls) and expose endpoints for integration with observability stacks (e.g., Prometheus).
   - Offer interactive control for step-through debugging and human-in-the-loop development.
   - Invest in documentation, guides, and a “cookbook” of agent scenarios to accelerate adoption and success.

## Industry Context & Sources

- See AGENTS.md for additional references.
- These recommendations ensure Agentry is positioned for reliability, scalability, security, and developer delight as the field evolves.

---

# === MASTER PROMPT: BUILDING "AGENTRY CLOUD" ===

#

# ░█▀█░█▀▀░█▀▀░█▀█░▀█▀░█▀▄░█░█

# ░█▀█░█░█░█▀▀░█░█░░█░░█▀▄░░█░

# ░▀░▀░▀▀▀░▀▀▀░▀░▀░░▀░░▀░▀░░▀░

#

# You are Codex, acting as a senior Go/cloud architect & engineer team.

# Implement the next‑generation “Agentry Cloud” platform by extending

# the existing OSS repository github.com/marcodenic/agentry (Go 1.23).

#

# ================================================================

# 0. CURRENT CONTEXT (read‑only, do NOT re‑implement)

# ------------------------------------------------

# • Language/runtime ............. Go 1.23 (CLI, TUI, HTTP).

# • Core capabilities ............ agent run loop, rule‑based model

# routing, in‑memory convo memory + toy vector store, builtin tool

# registry (echo, bash, fetch, etc.), sub‑agent spawning, SSE/JSONL

# trace writer, TypeScript SDK.

# • Weaknesses ................... no persistence, no orchestration DSL,

# no sandbox, limited observability, single‑process only.

#

# ================================================================

# 1. HIGH‑LEVEL GOALS

# ------------------------------------------------

# G1. **Cloud‑Ready Multi‑Agent Runtime** – agents can run, pause, resume,

# scale horizontally, and collaborate across processes/nodes.

# G2. **Self‑Building Platform** – agents eventually bootstrap, test,

# deploy, and improve Agentry itself with minimal human input.

# G3. **Enterprise‑Grade Safety & Observability** – sandboxing, RBAC,

# fine‑grained tool permissions, real‑time dashboards, metrics.

#

# ================================================================

# 2. MAJOR EPICS & FUNCTIONAL REQUIREMENTS

# ------------------------------------------------

# E1. **Persistent Memory & Resumable Workflows**

# R1.1 Pluggable durable KV (+ optional Postgres) for short texts.

# R1.2 Pluggable vector DB (e.g. SQLite‑vss, Milvus, pgvector).

# R1.3 `core.Agent.SaveState(id)` / `LoadState(id)` snapshots:

# • messages, tool call history

# • custom user state blobs

# R1.4 Resume flag in CLI & HTTP (“continue run <id>”).

#

# E2. **Declarative YAML Workflow DSL**

# R2.1 Extend `.agentry.yaml` → `.agentry.flow.yaml`.

# R2.2 Grammar primitives: `agents`, `tasks`, `parallel`, `sequential`,

# `condition`, `loop`, `timeout`, `env`.

# R2.3 Engine that interprets DSL → spawns/coordinates agents. (basic engine implemented)

# R2.4 Validation + static schema (JSON‑Schema for IDE support).

#

# E3. **Sandboxing & Secure Tool Execution**

# R3.1 Default “safe mode”: all shell/code tools run in Docker or

# gVisor micro‑VM; host path `/workspace` mounted read‑write.

# R3.2 Per‑tool manifest flags: `privileged: false`, `net: [host...]`,

# `cpu_limit`, `mem_limit`.

# R3.3 Interactive & policy‑file approval for privileged actions.

#

# E4. **Plugin API & Marketplace**

# R4.1 Go interface `plugin.Tool` already exists; formalize semver,

# register via `init()`.

# R4.2 External HTTP‑based tools: OpenAPI / MCP descriptors allowed.

# R4.3 `agentry plugin install github.com/foo/bar` – pulls source,

# `go install`, updates manifest.

#

# E5. **Distributed Scheduling & Collaboration**

# R5.1 gRPC service `AgentHub`:

# • `Spawn(SubAgentRequest) → AgentID`

# • `SendMessage(AgentID, Input) → Ack`

# • server‑side stream `Trace(AgentID) → TraceEvent`

# R5.2 Simple gossip or etcd‑based registry for node discovery.

# R5.3 Round‑robin & binpack schedulers; pluggable strategy.

# R5.4 CLI `agentry node join --hub <addr>`; Helm chart for k8s.

#

# E6. **Observability Suite**

# R6.1 Prometheus metrics endpoint (`/metrics`): token counts,

# tool‑latency histograms, goroutines, memory.

# R6.2 Web dashboard (Go + HTMX or SvelteKit) w/ panels:

# • Live agent timeline (“AI workflow debugger”)

# • Tool invocation log & arguments

# • Heat‑map of token/cost usage

# R6.3 Trace JSONL → Web UI viewer (drop‑file to replay).

#

# E7. **CI/CD & Self‑Improvement Loop**

# R7.1 GitHub Actions workflow: run `agentry eval`, `go test ./...`,

# `ts-sdk tests`, `golangci-lint`.

# R7.2 “DevOps Crew” flow file: manager‑agent → coder‑agent →

# reviewer‑agent → integration‑agent; on success pushes PR.

# R7.3 Canary deploy of new Agentry Docker image to fly.io/k8s.

#

# ================================================================

# 3. NON‑FUNCTIONAL REQUIREMENTS

# ------------------------------------------------

# • Performance: concurrent agent throughput ≥1k req/min on 4‑vCPU box.

# • Security:   sandbox escape zero, TLS for all gRPC calls, OPA for RBAC.

# • Reliability: snapshot & resume survives crash; 99.9 % uptime goal.

# • Extensibility: add new language SDKs w/o touching core.

#

# ================================================================

# 4. ARCHITECTURE / COMPONENTS

# ------------------------------------------------

# [A] Core Runtime (pkg/core) – existing loop, add snapshot logic.

# [B] Memory Service (memstore) – interface + SQLite / Postgres impl.

# [C] Flow Engine (pkg/flow) – parses YAML, orchestrates tasks.

# [D] Sandbox Executor (pkg/sbox) – wraps commands in Docker/gVisor.

# [E] Hub Node (cmd/agent‑hub) – gRPC server, scheduler, state DB.

# [F] Worker Node (cmd/agent‑node) – connects to Hub, executes agents.

# [G] Web Console (ui/web) – SPA served by Hub; uses SSE & REST.

# [H] CLI/TUI (cmd/agentry, tui/) – extended to control cluster.

#

# All internal comms → protobuf definitions in `api/*.proto`.

#

# ================================================================

# 5. DELIVERABLES & MILESTONES

# ------------------------------------------------

# M1 – Design docs + protobufs + Flow DSL schema.

# M2 – Persistent memstore + snapshot/resume + unit tests.

# M3 – Sandbox Executor MVP; tool permission flags.

# M4 – AgentHub & Worker Node; distributed demo (local).

# M5 – Observability metrics + Web dashboard alpha.

# M6 – End‑to‑end flow: DevOps Crew auto‑updates repo.

#

# ================================================================

# 6. ACCEPTANCE CRITERIA

# ------------------------------------------------

# • “Hello Cloud” demo: run `make demo-cloud` → spins 3 nodes, executes

# sample flow, dashboard shows live traces.

# • Crash‑and‑resume test: kill worker process mid‑task, restart, task

# completes correctly.

# • Security test: attempt forbidden `rm -rf /` in sandbox → blocked.

# • Load test: 1k concurrent `invoke` calls succeed < 200 ms p95.

#

# ================================================================

# 7. CODING GUIDELINES

# ------------------------------------------------

# • Go 1.23, modules tidy, `go test ./... -race` clean.

# • Use context.Context everywhere; no blocking ops without ctx.

# • Protobuf + buf.build; generate gRPC stubs with `buf generate`.

# • Integrate opentelemetry-go for traces (export to OTLP & stdout).

# • Commit messages Conventional Commits; PR titles semantic.

#

# ================================================================

# 8. EXECUTION INSTRUCTIONS FOR CODEX

# ------------------------------------------------

# 1. Work epic‑by‑epic; write/modify Go code, Dockerfiles, Helm chart,

# JS/TS for dashboard, docs in /docs.

# 2. After each epic: run `make test` → all Go + TS tests green.

# 3. When blocking, insert TODO comments & proceed; we’ll iterate.

# 4. Keep pull requests ≤1k LOC; reference spec section lines.

#

# ================================================================

# 9. OUT‑OF‑SCOPE (FOR NOW)

# ------------------------------------------------

# ✘ Full user‑facing SaaS auth/billing (future).

# ✘ Proprietary model fine‑tuning.

# ✘ Non‑Go core rewrite; JS SDK remains optional.

#

# ================================================================

# 10. BEGIN NOW

# ------------------------------------------------

# • First deliverable: README‑cloud.md describing new architecture +

# updated repo file tree. Include diagrams (ASCII or Mermaid).

# • Open a `design/` directory: add `flow_dsl.md`, `sandbox.md`,

# `distributed_arch.md`.

# • Open PR “feat(flow): initial DSL parser w/ tests”.

#

# >>> Proceed according to the spec above, asking clarifying questions

# ONLY when absolutely necessary. Output unified diffs or full file

# contents as needed. <END>

---

# 11. MACHINE‑READABLE TO‑DO MATRIX

> **Purpose:** A fully hierarchical, checkbox‑driven backlog for autonomous agents (or Codex) to execute without human interpretation.  
> Keep the core small; optional modules compile behind build tags or run side‑cars.

### Legend

- `‑ [ ]` unchecked work item – implement and open PR
- `ID` → unique reference; use in commit messages (`feat(mem): implement 1.1`)
- **Deps** column lists blocking IDs

---

## ❶ Persistent Memory & Workflow Resumption (core)

| ID  |  ‑ [ ] Task                                | Why              | How (high‑level)                           | Deps |
| --- | ------------------------------------------ | ---------------- | ------------------------------------------ | ---- |
| 1.1 | Pluggable `Store` back‑ends (file, SQLite) | Survive restarts | `StoreFactory` switch by YAML `memory:`    | —    | ✅
| 1.2 | ~~VectorStore → Qdrant/Faiss adapter~~         | Real ANN search  | REST or local lib via CGO                  | 1.1  |
| 1.3 | Checkpoint API (`Checkpoint()/Resume()`)   | Pause/continue   | Serialize loop state JSON after each event |  1.1 | ✅
| 1.4 | Max‑iteration graceful yield               | Avoid hard cap   | Emit `EventYield` when limit reached       |  1.3 | ✅
| 1.5 | ~~Session GC daemon~~                          | Disk hygiene     | TTL sweep & compaction                     |  1.1 | ✅

## ❷ Sandboxing & Security (security)

| ID  |  ‑ [ ] Task                 | Why                 | How                                  | Deps |
| --- | --------------------------- | ------------------- | ------------------------------------ | ---- |
| 2.1 | Tool permission matrix      | Block dangerous ops | YAML `permissions:` gate before exec | —    | ✅
| 2.2 | Docker sandbox executor     | Isolate shell/code  | Wrap tools in `docker run`           |  2.1 | ✅
| 2.3 | gVisor / Firecracker runner | Hard isolation      | CRI shim prototype                   |  2.2 | ✅
| 2.4 | Signed‑plugin registry      | Supply‑chain trust  | Publish index.json + SHA256 sig      | —    | ✅

## ❸ Declarative Workflow DSL (dx)

| ID  |  ‑ [ ] Task                 | Why                   | How                         | Deps |
| --- | --------------------------- | --------------------- | --------------------------- | ---- |
| 3.1 | YAML Flow schema v0.1       | Non‑dev orchestration | `.flow.yaml` → `[]Step`     | —    | ✅
| 3.2 | Flow runner engine          | Execute DSL           | Goroutines; aggregate trace | 3.1  | ✅
| 3.3 | CLI `agentry run flow.yaml` | DX                    | Detect & run                | 3.2  | ✅
| 3.4 | Sample flows repo           | Showcase              | 3 demos under `examples/`   | 3.2  | ✅

## ❹ Distributed Scheduling & Cluster Mode (cloud)

| ID  |  ‑ [ ] Task               | Why              | How                                   | Deps |
| --- | ------------------------- | ---------------- | ------------------------------------- | ---- |
| 4.1 | Agent UUID abstraction    | Decouple memory  | Assign UUID on spawn                  |  1.1 | ✅
| 4.2 | `/spawn` `/kill` HTTP API | Multi‑tenant     | REST endpoints map UUID→Store         |  4.1 | ✅
| 4.3 | NATS task queue           | Off‑process exec | Publish/subscribe `Task`              |  4.1 | ✅
| 4.4 | Worker micro‑service      | Run tasks        | `agentry worker --queue …`            |  4.3 | ✅
| 4.5 | Autoscaler                | Balance load     | Watch queue lag, scale k8s Deployment |  4.4 | ✅

## ❺ Agent Teams & Personas (dx)

| ID  |  ‑ [ ] Task           | Why                 | How                                    | Deps |
| --- | --------------------- | ------------------- | -------------------------------------- | ---- |
| 5.1 | Role‑template library | Plug‑and‑play crews | YAML snippets under `templates/roles/` |  3.1 | ✅
| 5.2 | Team presets          | Ready‑made crews    | Compose roles + flow                   |  5.1 | ✅
| 5.3 | Personality vars      | Tone control        | Mustache substitution                  |  5.1 | ✅
| 5.4 | Team‑chat TUI panes   | Visual demo         | Extend existing TUI                    |  5.2 | ✅

## ❻ Tooling & Plugin Ecosystem (community)

| ID  |  ‑ [ ] Task                    | Why             | How                           | Deps |
| --- | ------------------------------ | --------------- | ----------------------------- | ---- |
| 6.1 | `agentry tool init` scaffolder | Fast plugin dev | Gen skeleton + manifest       | —    | ✅
| 6.2 | Community registry site        | Discovery       | Static Jamstack index         |  6.1 | ✅
| 6.3 | Plugin installer CLI           | Easy add        | `install <repo>` updates YAML |  6.2 | ✅
| 6.4 | OpenAPI/MCP adapter            | Interop         | Wrap external spec as tool    |  6.3 | ✅

## ❼ Observability & Telemetry (ops)

| ID  |  ‑ [ ] Task         | Why       | How                    | Deps |
| --- | ------------------- | --------- | ---------------------- | ---- |
| 7.1 | OTLP trace exporter | Aggregate | Convert events → spans |  4.4 | ✅
| 7.2 | Prometheus metrics  | Ops       | `/metrics` endpoint    |  4.4 | ✅
| 7.3 | Web dashboard       | Visualize | Next.js or SvelteKit   |  7.1 | ✅
| 7.4 | Cost estimator      | Budget    | Post‑run analyzer      |  7.2 |
| 7.5 | Profiling dashboard | Perf tune | pprof + flamegraphs    |  7.2 |

## ❽ Automated Test & CI (qa)

| ID  |  ‑ [ ] Task            | Why              | How                            | Deps |
| --- | ---------------------- | ---------------- | ------------------------------ | ---- |
| 8.1 | Stress plan → Go tests | Regression       | `pkg/e2e` suite                |  1.3 | ✅
| 8.2 | GitHub Actions matrix  | Multi‑os         | linux/mac/win jobs             |  8.1 | ✅
| 8.3 | Chaos monkey job       | Crash resilience | Kill worker mid task           |  4.4 | ✅
| 8.4 | ~~Benchmarks~~             | Perf trend       | `go test -bench` + flamegraphs |  7.2 |

## ❾ Developer Experience (dx)

| ID  |  ‑ [ ] Task                 | Why                | How                 | Deps |
| --- | --------------------------- | ------------------ | ------------------- | ---- |
| 9.1 | One‑line installer          | Fast start         | Homebrew/scoop/deb  | —    | ✅
| 9.2 | ~~Docker‑compose mini‑cluster~~ | Showcase           | `docker compose up` |  4.4 |
| 9.3 | Helm chart                  | Enterprise         | k8s templates       |  4.4 | ✅
| 9.4 | VS Code extension           | Editor integration | SSE panel           |  7.1 | ✅
| 9.5 | Multi-agent TUI redesign    | Usability          | See TUI_IMPLEMENTATION_PLAN.md |  —  |
| 9.6 | TUI command system          | Control agents     | /spawn /stop /switch commands   | 9.5 | ✅ |
| 9.7 | Real-time agent dashboard   | Visual status      | spinners + token bars (in progress) | 9.5 |
| 9.8 | Custom theming              | Brand look         | lipgloss styles & presets | 9.5 |

## ❿ Documentation & Examples (docs)

| ID   |  ‑ [ ] Task        | Why             | How                        | Deps          |
| ---- | ------------------ | --------------- | -------------------------- | ------------- |
| 10.1 | User guide rewrite | Up‑to‑date docs | mkdocs site                | All           | ✅
| 10.2 | ~~“Power demos” repo~~ | Marketing       | DevOps flow, ETL, research |  Sections 3‑5 |
| 10.3 | Tutorial videos    | Onboarding      | Asciinema + Loom           |  9.1          |

## ⓫ Governance (meta)

| ID   |  ‑ [ ] Task              | Why           | How            |
| ---- | ------------------------ | ------------- | -------------- |
| 11.1 | CODEOWNERS + RFC process | PR discipline | Add files      |
| 11.2 | Monthly roadmap call     | Feedback loop | Zoom + minutes |
| 11.3 | Contributor CLA          | Legal         | CLA‑assistant  |

## ⓬ Performance & Benchmarking (perf)

| ID   |  ‑ [ ] Task              | Why            | How                        | Deps |
| ---- | ------------------------ | -------------- | -------------------------- | ---- |
| 12.1 | Go vs Python micro‑bench | Speed story    | Compare throughput         |  8.4 |
| 12.2 | 1k‑agent scale test      | Validate scale | k8s job, record tokens/sec |  4.4 |

---

### Continuous‑Integration Hooks

- Any code under `core/`, `engine/`, `tools/` triggers **full e2e** (8.1).
- Nightly **bench job** posts flamegraph to GitHub Pages (12.1, 12.2).
- **Security‑mode** tests run both with and without sandbox flag (2.2).

---

### Changelog

- 2025‑06‑24 – Removed unused vector interface from `pkg/memstore` to simplify
  the codebase.

---

> **Mantra:** _Keep the core fast and tiny; everything else is optional & pluggable._
> End of machine‑readable backlog.
