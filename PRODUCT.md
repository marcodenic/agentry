Here’s a **drop‑in replacement** for your `PRODUCT.md` that integrates the CLI/config changes:

* **YAML‑first configuration** with clear precedence
* **Minimal, reliable flags** (`--config`, `--set`, `--debug`, `--theme`)
* **TUI default** (no args) and **implicit run** (`agentry <prompt>`)
* **Robust parsing** (flags must come before prompt; non‑interspersed parsing; `--` sentinel supported)
* **Env vars deprecated** (keep only `AGENTRY_CONFIG` for CI)

---

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
* **TUI/CLI:** TUI default when no args; **implicit run** with `agentry <prompt>`; **minimal flags**; YAML‑first config.
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

## Configuration (YAML‑first)

**YAML is the source of truth.** Flags are small overrides; env vars are deprecated.

**Discovery & Precedence (highest → lowest)**

1. **CLI flags** (`--config`, `--set`, `--debug`, `--theme`)
2. **`--set key=value`** overrides (merge into loaded YAML; supports nested paths)
3. **`--config /path/to/.agentry.yaml`** (explicit file)
4. **Auto‑discover** first existing:

   * `./.agentry.yaml`
   * `$(git root)/.agentry.yaml`
   * `$XDG_CONFIG_HOME/agentry/config.yaml` or `~/.config/agentry/config.yaml`
5. **Built‑in defaults**

**Example `.agentry.yaml`**

```yaml
model:
  provider: anthropic
  name: claude-4
roles:
  agent0:
    tools: [spawn, gather, todo, tree, grep, view, run_tests, lsp_diagnostics]
  coder:
    tools: [tree, grep, view, patch, run_tests, lsp_diagnostics]
scheduler:
  max_in_flight: 2
  tpm_guard: true
tui:
  theme: dark
workspace:
  root: .
```

---

## CLI Usage

**Grammar**

```
agentry [GLOBAL_FLAGS] [SUBCOMMAND] [SUBCOMMAND_FLAGS] [--] [PROMPT...]
```

**Behavior**

* **No args** → launch **TUI**
* **No subcommand but PROMPT present** → **implicit run** (Agent 0 with that prompt)
* **With subcommand** → run that subcommand

**Flags must come before the prompt.**
Parsing is **non‑interspersed**: the first non‑flag token starts the prompt. Use `--` only if your prompt begins with `-`.

### Minimal Global Flags

```
--config path/to/.agentry.yaml
--set key=value             # may repeat; merges into YAML (e.g., --set tui.theme=light)
--debug                     # debug logging
--theme dark|light|auto     # quick TUI override
```

### Subcommands (optional)

```
agentry tui                 # force TUI
agentry run "<prompt>"      # explicit run
agentry refresh-models      # update model pricing/cache
agentry config doctor       # print merged config + sources
```

### Examples

```bash
# TUI (default)
agentry
agentry --theme dark

# Direct prompt (implicit run)
agentry fix the failing tests
agentry --debug fix the failing tests

# Flags before prompt, everything after first non-flag is the prompt
agentry --debug add a --force flag to the CLI help

# If your prompt must start with '-', use the sentinel:
agentry -- "--help me add --force without parsing as flags"

# Explicit subcommands
agentry run "add health check endpoint"
agentry tui
agentry refresh-models

# YAML overrides without new flags:
agentry --config ./my.yaml --set scheduler.max_in_flight=3 --set tui.theme=light "update CI"
```

### Environment Variables (Deprecated)

Env vars are deprecated; use YAML + flags. The only supported var is:

* `AGENTRY_CONFIG=/path/to/.agentry.yaml` — config file path (CI convenience)

When other legacy env vars are detected, print a one‑line deprecation notice and ignore.

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
12. **CLI hardening**

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
    * Switch parser to **non‑interspersed** mode; treat first non‑flag as prompt
    * Implement `implicit run` + `tui` default
    * Add `--set` key=value merges and **config doctor**
    * Add CLI golden tests for all examples above
    * Emit deprecation warnings for legacy env vars

---

## TODO — Prompt Engineering & Structuring

* [ ] **Prompt compiler**: render XML system prompts; minimal tag vocab (`<sop>`, `<tools>`, `<task-spec>`, `<running-summary>`, `<output-format>`).
* [ ] **Templates**: Agent 0, Coder, Tester, Critic SOPs; output JSON schemas; tool lists per role.
* [ ] **Escaping**: CDATA/escape all untrusted content; golden tests.
* [ ] **Caps**: enforce token caps for system message; keep outputs concise.
* [ ] **Validation**: prompt render unit tests; “no dangling tags”; JSON schema checks on outputs.
* [ ] **Renderer AB‑switch (later)**: keep XML/Markdown pluggable (experiments).
* [ ] (Optional) **10‑part structure** experiments for Agent 0 orchestration; keep disabled by default to preserve Context‑Lite.

---

## BUGS

* On window resize: “malformed char codes”.
* No `reasoning_effort` support.

---

**Update Policy:** After material change, update this file + role templates + CLI help. Remove shipped items; avoid stale duplication.
**Status Legend:** Internal hardening stays until merged; user‑visible items move to “Recently Completed” once the minimal slice is shipped & documented.

*Historical PLAN.md & FEATURES.md merged here (updated 2025‑09‑06).*
