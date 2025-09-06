
# Agentry Product & Roadmap

Single authoritative doc. Keep terse, actionable. Update after each merge/re‑prioritization.

FOR AGENTS, run: `./agentry <prompt>`

---

## Vision (Condensed)

Local‑first, observable, resilient **multi‑agent** development orchestrator. Open any repo, point **Agent 0** at a planning doc or task list, and it **plans → delegates → implements → tests → reviews → finalizes (PR/commit)** with clear traces.

**Principles**

* **Context‑Lite:** inject only **SOP, TaskSpec, RunningSummary**. Agents fetch everything else via tools.
* **Agent 0 orchestrates:** one spawn path; runtime scheduler handles concurrency (no special “parallel tool”).
* **Quality gate:** a task is *done* only when tests are green and a Critic approves (if relevant to task).
* **Durable memory:** per‑agent disk history with short RunningSummary in prompts.
* **Terminal‑first UX:** tasks, agents, events, artifacts visible as they happen.
* **No per‑agent inbox:** coordination is via Agent 0, the TODO store, and workspace events (shared log).

---

## Current Foundations (What Exists)

* **Core loop:** tool calling, streaming, tracing, error‑as‑data, retry caps.
* **Tools:** 30+ built‑ins (atomic file ops, search/replace, web/network, OpenAPI/MCP, audit/patch, delegation/spawn).
* **Models:** OpenAI + Anthropic via unified `model.Client` (streaming; usage tracked).
* **Multi‑agent:** team registry + delegation; Agent 0 role = orchestrator (spawn/manage workers).
* **Memory:** per‑agent convo history + vector store; SharedStore (mem/file); basic checkpointing.
* **Coordination:** **workspace events** feed (shared), **TODO store** (planning memory). **Per‑agent inbox removed.**
* **TUI:** live stream, delegation events, token/cost bar, diagnostics summary, safe autoscroll.
* **CLI:** direct prompts (+ default TUI when no command) and `refresh-models`. Legacy commands removed.
* **Context:** **minimal builder** in place; **Context‑Lite** compiler incoming (replacing Context v2).

---

## Recently Completed (Highlights)

* SharedStore (mem+file) with TTL/GC; persist coordination events.
* Delegation safety: worker agents lose `agent` tool.
* LSP diagnostics surfaced in TUI (gopls / tsc).
* Minimal context builder shipped; heavy hardcoded text removed.
* Pricing cache path moved to user cache dir; `refresh-models` command available.
* Iteration cap removed (agent runs to final) with optional budget stop.
* **Sprint Complete: Major Cleanup & Simplification**

  * Removed Prometheus/metrics system completely (code, deps, configs)
  * Removed eval system entirely (internal/eval, commands, test files)
  * Simplified CLI to core functionality; eliminated legacy commands
  * Agent 0 TUI display fixed (shows “Agent 0” with “System” role)
  * All tests passing; cleaner, lighter codebase

---

## Hardening & Cleanup (No New UX; ship fast)

### Architecture

* [ ] **Context‑Lite Prompt Compiler** (XML system prompt; inject **SOP + TaskSpec + RunningSummary**; outputs JSON).
* [ ] Introduce `AgentConfig` (budgets, error handling, model name) to reduce env sprawl.
* [x] Extract tool execution from `Agent.Run` → `executeToolCalls` (smaller CC, testable).
* [x] Cancellation checks pre/post model call & per tool.
* [x] Fallback minimal system prompt if role file missing.

### Code Quality

* [x] Consolidate env helpers into `internal/env`.
* [x] Replace O(n²) compaction sort with `sort.Slice`.
* [ ] Clarify spawn semantics (shared vs isolated vector store) + option toggle.
* [ ] Collapse duplicate default prompt helpers; keep one public API.
* [ ] Guard verbose dumps behind selective debug channels.
* [ ] Normalize model names (`provider/model`) via helper.

### **Remove / Retire (this sprint)**

* [ ] **Remove Context v2 pipeline** code & configs. Delete provider‑based relevance/budget assembly.
* [ ] **Remove `parallel_agents` (or any parallel tool path)** — consolidate on single `spawn/gather`; runtime scheduler handles concurrency.
* [ ] **Remove auto “related files”/vector sweeps** from prompt assembly. Retrieval happens via tools only.
* [ ] **Remove per‑agent inbox messaging** — delete `send_message`, `inbox_read`, `inbox_clear`, `request_help`; remove “INBOX CONTEXT” injection; delete 📬/🆘 console prints; migrate signals to TODOs or workspace events.
* [ ] **Remove pinned‑rules block** from prompts/configs; move any global guidance into **role SOPs** and **runtime enforcement** (output JSON validation, echo guard).
* [ ] Mark `AGENTRY_DISABLE_CONTEXT` **deprecated/no‑op** (pipeline removed).

### Testing

* [ ] Unit: tool error recovery (consecutive cap), history compaction edges, spawn inheritance deep copy.
* [ ] Golden: **XML prompt rendering** (token cap, escaping/CDATA, tool lists, no dangling tags).
* [ ] Integration: JSON stdout purity regression; spawn/gather with queued execution under artificial TPM.
* [ ] Regression: ensure no inbox or pinned‑rules injection paths remain; workspace events continue to surface team signals.

### Docs

* [ ] CONTRIBUTING: layers, adding tool/provider, test matrix, release steps.
* [ ] Memory architecture diagram (conversation vs vector vs shared store).
* [ ] Role authoring guide (SOPs, tool allowlists, output schemas).
* [ ] Remove inbox & pinned‑rules references from docs; add “coordination via TODO + workspace events”.

---

## Priorities (User‑Visible)

### High

1. **Context‑Lite Prompt Compiler + XML prompt bodies**

   * Minimal injection; CDATA/escape; golden tests; outputs JSON.
2. **Role SOPs (Agent 0, Coder, Tester, Critic)**

   * Standardize allowed tools + output schemas; concise, stepwise SOPs.
3. **Agent TODO & Planning Memory**

   * Persistent TODOs (CRUD, filters, comments, attachments) + TUI board.
4. **Spawn/Gather + Scheduler**

   * Agent 0 spawns multiple workers when independent; runtime enforces TPM queues; TUI agents panel.
5. **Memory & RunningSummary (per agent)**

   * Disk logs; thresholded summarization to short RunningSummary injected in prompts.
6. **QA Loop (Tests, LSP, Critic) — enforced**

   * Coder must run tests + LSP; Critic must approve before DONE.
7. **Auto‑LSP Loop**

   * Diagnostics post‑edit; TUI diag panel; feed into next turn.

### Medium

* **AST‑Based Editing v1** (Go/TS/JS) + formatter/diag validation + fallback.
* TUI polish: spinner, unified stream tail, syntax highlighting; agent cycling keybind fixes; Nerd Font optional glyphs.
* Planning‑doc ingestion (parse PRODUCT.md/ROADMAP/TASKS.md → TODOs).
* Normalize model names; spawn semantics toggle; collapse default prompt helpers.

### Deferred / Later

* **Cost Accuracy Pass** (usage parsing, pricing loader, TUI totals/budgets).
* Advanced checkpointing; resumable workflows UX.
* Multi‑provider plugins; remote/cluster spawn; event bus; distributed teams.
* Guardrail frameworks; sandbox hardening.

---

## Roadmap (Milestones)

### **M1 — Context‑Lite & SOPs (Now)**

* [ ] Remove **Context v2** pipeline from code and docs.
* [ ] Implement **Prompt Compiler (Context‑Lite)** → XML system message from role config + TaskSpec + RunningSummary.
* [ ] Author SOPs + output JSON schemas for **Agent 0** and **Coder**; update role configs & tool allowlists.
* [ ] Golden tests: render stability, token caps, escaping/CDATA.

### **M2 — TODO Store, Memory & Scheduler**

* [ ] TODO tool + persistence under `.agentry/` (CRUD, comments, attachments).
* [ ] Per‑agent history logs + RunningSummary on threshold.
* [ ] `spawn` + `gather` API; TPM‑aware scheduler; basic TUI Agents panel.

### **M3 — QA Loop**

* [ ] Coder enforces tests + LSP before returning; surfaces diffs + commit message.
* [ ] Tester returns concise failures or “✅ Tests passing”.
* [ ] Critic gate; Agent 0 only marks DONE when tests green + critic approve.
* [ ] TUI panels for tests/LSP/critic.

### **M4 — Auto‑LSP & Tests Integration**

* [ ] Language detection; run diagnostics automatically after writes; dedupe per tick.
* [ ] Test‑runner detection (`go test` / `npm test` / `pytest`); run after builds; feed failures.

### **M5 — AST Editing v1**

* [ ] tree‑sitter / native parsers (Go/TS/JS): `rename_symbol`, `replace_by_query`, `ensure_import`, `apply_patch_tree`.
* [ ] Fallback to line edits when AST fails; validate with formatter/linter + diagnostics.

### **M6 — UX Polish & Docs**

* [ ] Diff preview; syntax highlighting; spinner/stream improvements.
* [ ] CONTRIBUTING; memory architecture; examples.

### **M7 — Cost Accuracy (Later)**

* [ ] Accurate usage parsing; pricing loader; TUI totals; budgets (warn/stop).

---

## Context‑Lite Prompt Compiler (Summary)

* **Inject only:** **SOP, TaskSpec, RunningSummary**.
* **Prompt body:** XML tags; config remains YAML/JSON; **outputs must be JSON**.
* **Security:** escape or wrap untrusted content in `<![CDATA[ ... ]]>`.
* **Token cap:** target ≤ \~1–1.5k input tokens per system message.
* **No providers:** remove provider‑based context assembly (RelatedFiles, WorkspaceSummary, GitDiff, TestFailures, LSPDefs, Memory, etc.). Agents retrieve via tools.

---

## Agent TODO & Planning Memory (Spec)

**Namespace:** `todo:project:<hash>`
**Item:** `id, title, description, acceptance, owner(agent-id|role), status(todo|wip|done), created_at, updated_at, depends_on[], tags[]`
**APIs:** `todo.add`, `todo.list`, `todo.update`, `todo.comment`, `todo.attach`
**Persistence:** file/SQLite under `.agentry/`
**TUI:** columns **TODO/WIP/DONE**, owner badges, acceptance chips.

---

## Spawn/Gather & Scheduler (Spec)

* `spawn(role, input, {todo_id?, artifacts?}) -> spawn_id`
* `gather(spawn_id | spawn_ids[]) -> results`

**Runtime scheduler**: TPM‑aware queues, max in‑flight per provider, fair ordering; aggregates results for Agent 0.
**Agent 0 SOP**: may issue multiple `spawn` for independent tasks; runtime handles concurrency.

---

## Memory & RunningSummary (Spec)

* Per‑agent logs: `.agentry/sessions/{session}/{agent}.log`.
* On threshold, compress oldest segment → **RunningSummary** (≈150–300 tokens) and replace dropped segment with pointer “See summary vN”.
* **Only** RunningSummary is injected in prompts; full logs remain on disk.

---

## QA Loop (Spec)

* **Coder** must run `run_tests` + `lsp_diagnostics`; iterate until green; return diff + proposed commit.
* **Tester** returns failures as `file:line — message` bullets or “✅ Tests passing”.
* **Critic** checks diff vs acceptance; outputs “✅ Approve” or blockers/nits.
* **Agent 0** marks TODO **DONE** only when tests green **and** Critic approves (or explicit user override).

---

## Auto‑LSP Loop (Spec)

* Start servers as needed; cache per workspace.
* Trigger after writes; batch per tick to avoid thrash.
* Surface errors/warnings per file; feed key diagnostics into next turn.

---

## AST‑Based Editing v1 (Spec)

* **Languages:** Go, TypeScript/JavaScript (Python optional).
* **Ops:** `rename_symbol`, `replace_by_query`, `ensure_import`, `apply_patch_tree`.
* **Validate:** run formatter/linter; auto diagnostics.
* **Fallback:** degrade to line edits if AST parse fails.

---

## Agent 0 Orchestration Loop (Runtime + Role Addendum)

1. **Plan** (read PRODUCT/ROADMAP; create/update TODOs with acceptance).
2. **Delegate** (spawn coder/tester/critic; independent tasks can proceed concurrently).
3. **Build** (Coder edits).
4. **Test** (auto detect & run tests; capture failures).
5. **Review/Critic** (approve or blockers).
6. **Integrate** (summarize; propose commit/PR text).
7. **Verify‑Done** (re‑run tests; close TODO).
8. **Iterate** as needed.

---

## CLI Usage

Agentry supports four main commands:

### Primary Usage

```bash
# Direct prompt execution (Agent 0 responds directly)
agentry hello there
agentry fix the failing tests
agentry implement a health check endpoint

# Use quotes when you need special characters
agentry "analyze the codebase structure & suggest improvements"
agentry "implement user auth with JWT tokens & refresh logic"

# Start TUI interface (default when no command)
agentry
agentry --config custom.yaml

# Update model pricing data
agentry refresh-models

# Show help
agentry help
```

### Common Flags

```bash
--config path/to/.agentry.yaml
--theme dark|light|auto
--debug
--keybinds path/to/keybinds.json
--creds path/to/creds.json
--mcp server1,server2
--save-id session1
--resume-id session1
--checkpoint-id ckpt1
--port 8080
```

### Examples

```bash
agentry list all TODO comments in the codebase
agentry fix the failing tests
agentry add error handling to the auth module

agentry --debug analyze the codebase structure

agentry --resume-id my-session --theme light

agentry "implement user auth with JWT tokens & refresh logic"
agentry "analyze performance bottlenecks using profiling data"

agentry refresh-models
```

### Environment Variables (Alternative Configuration)

**Core Settings**

* `AGENTRY_DEBUG=1` — Enable debug output (prefer `--debug`)
* `AGENTRY_THEME=dark` — Set theme (prefer `--theme`)
* `AGENTRY_ENV_FILE=/path/to/.env` — Load env from file

**Advanced/Internal**

* `AGENTRY_TUI_MODE=1` — Internal flag (set automatically)
* `AGENTRY_DEFAULT_PROMPT="..."` — Default prompt override
* `AGENTRY_AUDIT_LOG=/path/to/audit.log` — Audit logging
* `AGENTRY_HISTORY_LIMIT=100` — Chat history limit
* `AGENTRY_DELEGATION_TIMEOUT=300` — Delegation timeout
* `AGENTRY_MODELS_CACHE=/path/to/cache` — Model cache location
* `AGENTRY_STORE_GC_SEC=3600` — Memory store GC interval

**Tool/Filter Controls (Deprecated — use CLI flags)**

* ~~`AGENTRY_DISABLE_TOOL_FILTER=1`~~ — use `--disable-tools`
* ~~`AGENTRY_TOOL_ALLOW_EXTRA=tool1,tool2`~~ — use `--allow-tools`
* ~~`AGENTRY_TOOL_DENY=tool1,tool2`~~ — use `--deny-tools`
* `AGENTRY_DISABLE_CONTEXT=1` — **deprecated/no‑op** (Context v2 removed)

**Context/Memory Tuning**

* `AGENTRY_CTX_CAP_AGENT0=8000` — Context cap for Agent 0
* `AGENTRY_CTX_CAP_WORKER=4000` — Context cap for workers

**Logging/Communication**

* `AGENTRY_COMM_LOG=1` — Enable comm logging
* `AGENTRY_COLLECTOR=...` — Telemetry endpoint
* `AGENTRY_PORT=8080` — HTTP server port (prefer `--port`)

### Migration from Environment Variables

```bash
# Old
AGENTRY_DEBUG=1 AGENTRY_THEME=dark agentry "task"

# New
agentry --debug --theme dark "task"
```

---

## Next Steps (Tight List)

1. **Delete** Context v2 pipeline (+ providers, docs); remove auto related‑files/vector sweeps.
2. **Delete** per‑agent inbox messaging (tools + injection + prints); migrate “help/notify” flows to TODOs or workspace events.
3. **Delete** pinned‑rules injection; move any remaining policy lines into SOPs; add runtime **JSON output validation** and **echo guard**.
4. **Implement** Context‑Lite Prompt Compiler (XML body; JSON outputs; CDATA/escape; golden tests).
5. **Author** SOP prompts (Agent 0, Coder); update role configs + tool allowlists.
6. **Build** TODO tool + persistence + TUI board.
7. **Add** per‑agent history + RunningSummary (thresholded).
8. **Ship** spawn/gather & scheduler; TUI Agents panel.
9. **Wire** QA loop (tests + LSP + critic) & enforce **DONE** gate.
10. **Add** Auto‑LSP post‑edit with TUI panel.
11. **Prepare** AST v1 ops (Go/TS/JS) with validation + fallback.

---

## TODO - Prompt Engineering & Structuring

### Prompt Structure Improvements (XML, but order the elements in this sequence where relevant)
- [ ] **Implement structured prompt framework** following the 10-part structure:
  1. Task context
  2. Tone context  
  3. Background data, documents, and images
  4. Detailed task description & rules
  5. Examples
  6. Conversation history
  7. Immediate task description or request
  8. Thinking step by step / take a deep breath
  9. Output formatting
  10. Prefilled response (if any)
- [ ] **Apply prompt structure to Agent 0 orchestration** - ensure planning/delegation prompts follow structured format
- [ ] **Standardize worker agent prompts** with consistent structure across coder, tester, reviewer roles
- [ ] **Add prompt template validation** to ensure all agent prompts follow the structured framework
## BUGS

* On window resize: “malformed char codes”.
* No `reasoning_effort` support.

---

## TODO — Prompt Engineering & Structuring

* [ ] **Prompt compiler**: render XML system prompts; minimal tag vocab (`<sop>`, `<tools>`, `<task-spec>`, `<running-summary>`, `<output-format>`).
* [ ] **Templates**: Agent 0, Coder, Tester, Critic SOPs; output JSON schemas; tool lists per role.
* [ ] **Escaping**: CDATA/escape all untrusted content; golden tests.
* [ ] **Caps**: enforce token caps for system message; keep outputs concise.
* [ ] **Validation**: prompt render unit tests; “no dangling tags”; JSON schema checks on outputs.
* [ ] **AB‑switch (later)**: keep renderer pluggable (XML vs Markdown) for experiments.
* [ ] (Optional) **10‑part structure** experiments for Agent 0 orchestration; keep disabled by default to preserve Context‑Lite.

---

**Update Policy:** After material change, update this file + role templates + CLI help. Remove shipped items; avoid stale duplication.
**Status Legend:** Internal hardening stays until merged; user‑visible items move to “Recently Completed” once the minimal slice is shipped & documented.

*Historical PLAN.md & FEATURES.md merged here (updated 2025‑09‑06).*
