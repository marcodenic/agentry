# Agentry Architectural Plan & Roadmap

This document outlines the architectural evolution of Agentry, focusing on scalability, reliability, and non-interactive capabilities. It builds upon the vision in `PRODUCT.md` with specific technical recommendations.

## 1. Core Architecture: The "Mesh" & "Async" Evolution

To move beyond simple request-response chains and enable true multi-agent collaboration, we must evolve the core orchestration primitives.

### A. Identity-Aware Delegation (The "Mesh")
**Current State:** `Team.Call` hardcodes "agent_0" as the source.
**Problem:** Peer-to-peer delegation (e.g., Manager -> Coder) is logged as Agent 0 -> Coder, losing the chain of command.
**Plan:**
- [ ] Update `Team.Call` signature to `Call(ctx, fromAgent, toAgent, input)`.
- [ ] Update `agent` tool to inject the *current* agent's identity into the call.
- [ ] Ensure `CoordinationEvent` logs reflect the true `from` -> `to` relationship.

### B. Async Job System (The "Non-Blocking" Orchestrator)
**Current State:** `Team.Call` blocks until the sub-agent finishes.
**Problem:** Long-running tasks (refactors, test suites) freeze the orchestrator.
**Plan:**
- [ ] Implement `JobStore` (in-memory + persistence) to track task state.
- [ ] Create `spawn_job(role, input) -> job_id` tool.
    - Returns immediately with a Job ID.
    - Agent 0 can continue planning or answering user queries.
- [ ] Create `check_job(job_id)` and `wait_for_job(job_id)` tools.
- [ ] Implement `JobScheduler` to manage concurrency limits (e.g., max 2 coders at once).

### C. "Pull-Based" Context (The Blackboard Pattern)
**Current State:** `workspaceContext` is string-concatenated and pushed into every prompt.
**Problem:** Context window pollution; agents get irrelevant noise.
**Plan:**
- [ ] **Structured Shared State:** Define a schema for Project State (Current Phase, Active Blockers, Recent Major Events).
- [ ] **Pull Tools:** Give agents tools to query this state:
    - `read_project_state()`
    - `read_recent_events(limit=10)`
    - `read_agent_status(agent_id)`
- [ ] **Context-Lite:** Only inject the absolute minimum (Identity, Current Task, immediate "Inbox" notifications) into the system prompt. Everything else is fetched on demand.

---

## 2. Reliability & "Headless" Operation

Enabling Agentry to run reliably without human intervention (CI/CD, self-improvement loops).

### A. Non-Interactive / Headless Mode
**Current State:** CLI supports `agentry "prompt"`, but output is unstructured text.
**Plan:**
- [ ] **JSON Output Mode:** Add `--json` flag to CLI.
    - Output the final result, cost, and status as a JSON object.
    - Essential for programmatic consumption by other scripts or agents.
- [ ] **Exit Codes:** Ensure distinct exit codes for:
    - Success (0)
    - Task Failed (1)
    - Internal Error (2)
    - Budget Exceeded (3)
- [ ] **Quiet Mode:** Add `--quiet` or `-q` to suppress all TUI/spinner output, printing only the final result to stdout.

### B. Self-Correction Loops
**Current State:** Agents can retry on tool errors, but high-level task failure is often final.
**Plan:**
- [ ] **The "Critic" Loop:** Enforce a review step for *all* code changes in headless mode.
    - If `agentry "fix bug"` is run, the agent must self-verify (run tests) before exiting.
- [ ] **Post-Mortem Dump:** On failure, dump a `trace.json` that another agent can analyze to suggest fixes.

---

## 3. Implementation Roadmap

### Phase 1: Foundation (Identity & Async)
1. Refactor `Team.Call` to support `fromAgent`.
2. Implement basic `JobStore` and `spawn`/`check` tools.
3. Update `agent` tool to support both blocking (`call`) and non-blocking (`spawn`) modes.

### Phase 2: Context & State
1. Define `SharedState` schema.
2. Implement `read_state` tools.
3. Refactor System Prompt to remove "Push" context.

### Phase 3: Headless Hardening
1. Implement `--json` and `--quiet` flags.
2. Create a "Self-Test" suite where Agentry fixes its own bugs in a sandbox.

---

## 4. Known Issues & Immediate Fixes

### TUI Compilation & Integration (Critical)
- **Missing Field:** `viewState` struct lacks `ActivityFeed` field.
- **Initialization:** `ActivityFeed` component is not initialized in `NewWithConfig`.
- **Rendering:** `ActivityFeed` is not currently integrated into the main `View()` loop.

### Core Issues
- **`ls` Tool Sandbox:** The `ls` tool fails in some environments.
    - *Fix:* Implement a robust `list_files` fallback or ensure `ls` uses a safe, portable implementation (e.g., `os.ReadDir` instead of shell out).
- **Delegation Timeout:** Hardcoded 15m timeout.
    - *Fix:* Allow `spawn` tool to accept an optional `timeout` parameter.