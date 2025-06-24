# VS Code Extension

This extension streams agent output from a running Agentry server using the VS Code extension API and Server-Sent Events (SSE).

## Setup

```bash
cd extensions/vscode-agentry
npm install
npm run build
```

Use the VS Code Extensions view to load this folder as an extension ("Run Extension" from the debug tab or `code --extensionDevelopmentPath=...`).

## Connecting

1. Launch an Agentry server:
   ```bash
   agentry serve --config examples/.agentry.yaml
   ```
2. In VS Code, execute the **Agentry: Open Panel** command.
3. Enter the server URL (`http://localhost:8080` by default) and a message.
4. Watch the streaming tokens appear in the **Agentry SSE** output channel.
5. Run **Agentry: Stop Stream** to disconnect.

## Publishing

Use [vsce](https://github.com/microsoft/vsce) to package and upload the extension.

```bash
cd extensions/vscode-agentry
npm install
npm run package  # produces .vsix file
npm run publish  # requires publisher token
```

Update the version in `package.json` before publishing.
