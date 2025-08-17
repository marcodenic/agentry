## Feature / Improvement Tracker

This file tracks user‑visible UX and orchestration improvements. Items are moved to "Completed" once a minimal implementation ships. Remaining work / polish stays under "Planned / In Progress".

### Completed
- Scroll friendly history (no forced jump while streaming): viewport only autoscrolls when user was at bottom (`model_tokens.go` uses `AtBottom()` safeguard). Partial truncation now only occurs when `AGENTRY_HISTORY_LIMIT` is set; older automatic truncation complaint addressed.
- Input prompt history (Up/Down arrows) implemented (`model_keys.go` maintains `inputHistory`).
- Basic agent delegation safety: spawned worker agents have the `agent` tool removed to prevent delegation cascades (`team.Add` / `AddAgent`).
- Streaming token progress & usage bar (live token counting + progress percent in `model_tokens.go`).
- Minimal context builder sentinel (`buildContextMinimal`) reduces previous heavy static context; dynamic project summaries cached with TTL.

### Planned / In Progress
- Elapsed time counter per agent (sidebar timer / stopwatch) – not yet implemented.
- Startup diagnostics banner: show presence of Agent 0 prompt file + detected API keys right under logo.
- Cursor navigation: reserve Left/Right strictly for cursor; adopt `shift+left/right` (or configurable keys) for agent cycling (currently still using arrow keys for agent switches when input not focused).
- Output jitter while typing: minor layout shift still occasionally observed; needs investigation (likely due to dynamic width / progress bar updates).
- Code syntax highlighting / markup coloring in TUI (needs style/theme + code block detection pass).
- Nerd Fonts glyph support (optional; make fallback safe in terminals without NF) – evaluate footprint.
- Rich elapsed + status line combining spinner, elapsed, tokens, and error count.
- Context management v2: Provider → Budget → Assembler pipeline (see `PRODUCT.md`) replacing minimal builder.
- Agent tool dedupe: prevent multiple `agent` tool calls in a single iteration (design ready).

### Deferred / Nice To Have
- Themed adaptive progress bars per model cost tier.
- Inline diff viewer for patch previews before apply.
- Pluggable renderer for color‑blind friendly palettes.

### Notes
- Keep feature descriptions concise; link to issues once public tracker is enabled.
- When marking an item complete, add a short justification pointing to file / function.
