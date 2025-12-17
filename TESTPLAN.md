# Agentry Test Plan & Roadmap

This document outlines the strategy for ensuring Agentry's reliability, stability, and correctness. It serves as a guide for both human developers and the agent itself to verify system health.

## 1. Testing Layers

### A. Unit Tests (Go)
**Scope:** Individual functions, tools, and core logic.
**Location:** `*_test.go` files alongside source code.
**Command:** `go test ./...`

**Key Areas:**
- **Tools:** Verify each tool (e.g., `grep`, `patch`, `write`) behaves correctly in isolation.
- **Context:** Ensure prompt assembly respects token limits and formatting.
- **Cost:** Verify token counting and budget enforcement.
- **Team:** Test delegation logic, role loading, and event logging.

### B. Integration Tests (Headless/CLI)
**Scope:** End-to-end flows running via the CLI without TUI.
**Location:** `tests/` directory.
**Command:** `./scripts/test.sh` (or specific test scripts).

**Key Scenarios:**
- **Agent 0 Boot:** Can Agent 0 start and answer a simple "hello"?
- **Delegation:** Can Agent 0 spawn a Coder to write a file?
- **Tool Usage:** Can an agent use `ls`, `read`, and `write` in sequence?
- **Resilience:** Does the system recover from API errors or tool failures?

### C. TUI Manual Verification (Checklist)
**Scope:** Interactive elements that are hard to automate.
**Frequency:** Before every release.

**Checklist:**
- [ ] **Launch:** Run `agentry` (no args). Does TUI open?
- [ ] **Input:** Type a message. Does it wrap correctly?
- [ ] **Streaming:** Does the AI response stream token-by-token?
- [ ] **Tabs:** Can you switch between Chat (Tab) and Debug (Ctrl+D)?
- [ ] **Feed:** Can you toggle the Activity Feed (Ctrl+F)?
- [ ] **Resize:** Resize the terminal window. Does the layout adapt without corruption?
- [ ] **Quit:** Does Ctrl+C exit cleanly?

---

## 2. Self-Correction & "Headless" Testing

To enable Agentry to test itself, we will implement the following capabilities:

### A. The "Self-Test" Suite
A dedicated set of tasks that Agentry runs against itself.
**Command:** `agentry run --test-suite=self-check` (Future)

**Tasks:**
1.  "Create a file named `test_artifact.txt` with content 'success'."
2.  "Read `test_artifact.txt` and verify content."
3.  "Delete `test_artifact.txt`."

### B. JSON Output Mode
**Goal:** Allow programmatic verification of Agentry runs.
**Flag:** `--json`
**Output:**
```json
{
  "status": "success",
  "result": "I have completed the task.",
  "cost": 0.02,
  "trace_id": "12345"
}
```

### C. The "Critic" Loop
**Goal:** Enforce quality before "Done".
**Logic:**
If running in headless mode:
1.  Agent generates solution.
2.  System automatically spawns a "Critic" agent.
3.  Critic reviews diffs/output.
4.  If rejected, Agent retries.
5.  Only exit 0 if Critic approves.

---

## 3. Roadmap for Test Infrastructure

| Phase | Task | Status |
| :--- | :--- | :--- |
| **1** | **Fix TUI Compilation** | ✅ Done |
| **2** | **Fix `ls` Sandbox Issue** | 🚧 Pending |
| **3** | **Implement `--json` flag** | 📅 Next Sprint |
| **4** | **Create `tests/self_test.go`** | 📅 Next Sprint |
| **5** | **Automate TUI Testing (VHS/Expect)** | 🔮 Future |

## 4. How to Run Tests Now

```bash
# 1. Run all unit tests
go test ./...

# 2. Run specific integration test
go test ./tests/comprehensive_orchestration_test.go

# 3. Run the "Smoke Test" script
./scripts/test_single_tool.sh
```
