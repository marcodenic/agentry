# API

## Built-in Tools

Agentry ships with a collection of safe builtin tools. They become available when listed in your `.agentry.yaml` file:

```yaml
tools:
  - name: echo
    type: builtin
  - name: ping
    type: builtin
  - name: bash
    type: builtin
  - name: branch-tidy
    type: builtin
  - name: fetch
    type: builtin
  - name: glob
    type: builtin
  - name: grep
    type: builtin
  - name: ls
    type: builtin
  - name: view
    type: builtin
  - name: write
    type: builtin
  - name: edit
    type: builtin
  - name: patch
    type: builtin  - name: sourcegraph
    type: builtin
  - name: agent
    type: builtin
  - name: mcp
    type: builtin
```

The example configuration already lists these tools so they appear in the TUI's "Tools" panel.

### Environment Configuration

Copy `.env.example` to `.env.local` and fill in `OPENAI_KEY` to enable real OpenAI calls. The file is loaded automatically on startup and during tests.

To run evaluation with the real model:

```bash
OPENAI_KEY=your-key agentry eval --config my.agentry.yaml
```

If no key is present, the built-in mock model is used.

## HTTP Endpoints

When running `agentry serve`, the following JSON endpoints are available:

- **POST `/spawn`** – create a new agent from the `default` template. Returns
  `{ "agent_id": "<uuid>" }`.
- **POST `/kill`** – persist the agent's state and remove it from memory.
- **POST `/invoke`** – queue a message for an agent to process.
