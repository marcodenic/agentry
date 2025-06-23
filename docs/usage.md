# Usage

## Quick Start
```bash
# CLI dev REPL with tracing
go install github.com/marcodenic/agentry/cmd/agentry@latest
agentry dev

# HTTP server + JS client
agentry serve --config examples/.agentry.yaml
npm i @marcodenic/agentry
```
The `examples/.agentry.yaml` file contains a ready-to-use configuration for these commands.

You can now use subcommands instead of the --mode flag:
- `agentry dev` (REPL)
- `agentry serve` (HTTP server)
- `agentry tui` (TUI interface)
- `agentry eval` (evaluation)
- `agentry flow` (run `.agentry.flow.yaml`)

Example:
```bash
agentry flow .
```

Run the sample scenarios in `examples/flows`:
```bash
agentry flow examples/flows/research_task
agentry flow examples/flows/etl_pipeline
agentry flow examples/flows/multi_agent_chat
```

Pass `--resume-id name` to load a saved session and `--save-id name` to persist after each run.
Use `--checkpoint-id name` to continuously snapshot the run loop and resume after a crash.

### TUI Themes & Keybinds
Create a `theme.json` file to customise colours and keyboard shortcuts. Agentry looks for the file in the current directory and its parents, falling back to `$HOME/.config/agentry/theme.json`.
```json
{
  "userBarColor": "#00FF00",
  "aiBarColor": "#FF00FF",
  "keybinds": {
    "quit": "ctrl+c",
    "toggleTab": "tab",
    "submit": "enter"
  }
}
```
