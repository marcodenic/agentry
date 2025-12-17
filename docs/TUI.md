# TUI Guide

The terminal UI provides a focused interface for interacting with the main agent.

## Layout

- **Chat pane** – The main conversation stream showing agent narration, tool calls, and responses. Tool entries update in place while running so you can watch progress.
- **Status bar** – Shows token usage, cost tracking, and elapsed time at a glance.
- **Input area** – Type prompts and commands.

## Workflow

1. Type your request in the input area.
2. The agent streams its response and any tool calls appear inline.
3. Tool calls show status (running/done/error), duration, and target paths.
4. Reasoning content from models like o1/o3 appears with a 💭 indicator.

## Navigation Keys

| Key | Action |
|-----|--------|
| `Ctrl+N` / `Ctrl+P` | Navigate through history |
| `Home` / `End` | Jump to start/end of chat |
| `Ctrl+F` | Toggle follow mode (auto-scroll) |
| `Ctrl+H` | Toggle help display |
| `Ctrl+C` | Cancel current operation / Exit |
| `Tab` | Switch between panes |

## Slash Commands

Type `/help` inside the chat for available commands.

## Tips

- The chat is the source of truth for all tool output and execution history.
- Watch the status bar for token/cost tracking during long operations.
- Use `--debug` flag when launching to see verbose logging.
- For reasoning models, thinking content is throttled to prevent terminal overload.
