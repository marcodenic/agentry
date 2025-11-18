# TUI Delegation Cheatsheet

This UI keeps the conversation readable while still tracking what every delegated worker is doing.

## What Goes Where

- **Main chat (left)** – always shows the canonical stream of events. You will see:
  - Agent 0’s narration.
  - The delegation notice (`Delegating to coder…`).
  - Every tool call from Agent 0 and from any spawned agent, rendered inline just like the screenshot you shared (e.g. `✔ View ~/src/.../sidebar.go`). Those entries update in place while the tool runs so you can watch progress without changing focus.
- **Sidebar (right)** – one card per agent. Each card only carries a single status line that mirrors the latest tool event, so you can glance at what each worker is currently doing without clutter.

## Workflow

1. You instruct Agent 0.
2. If it delegates, the chat shows the delegation block immediately.
3. When the delegated agent calls a tool, the chat displays the tool line and keeps it updated (duration, path, etc.). The sidebar line for that agent mirrors the same text.
4. Once a tool finishes, both the chat entry and the sidebar line flip to `DONE` (or `ERROR` for failures). The agent card resets to idle when the final response comes back.

## Navigation Aids

- `ctrl+n` / `ctrl+p` – cycle through agent cards.
- `home` / `end` – jump to first/last agent.
- `ctrl+f` – focus/unfocus the activity log if you want to scroll through older tool entries (it stays hidden until there’s history).
- `ctrl+h` – collapse/expand the agent list for more room when needed.

## Tips

- The chat is the source of truth for tool output—use the sidebar as a quick status glance.
- Because every agent writes into the chat, you can copy the execution history straight from one place.
- When an agent seems stuck, look at its sidebar line: it shows the exact tool + target it’s busy with and updates as soon as something changes.

That’s the intended flow: chat-first visibility, with the sidebar acting as a lightweight dashboard so you never miss what the delegated workers are doing.
