# Agentry VS Code Extension

This extension streams agent output from a running Agentry server using Server‑Sent Events (SSE).

## Building

```bash
cd extensions/vscode-agentry
npm install
npm run build
```

## Usage

1. Start the Agentry server in a terminal:
   ```bash
   agentry serve --config examples/.agentry.yaml
   ```
2. In VS Code, run the **Agentry: Open Panel** command.
3. Enter the server URL (defaults to `http://localhost:8080`) and a message to send.
4. Streaming output appears in the **Agentry SSE** output channel.
5. Run **Agentry: Stop Stream** to close the connection.
