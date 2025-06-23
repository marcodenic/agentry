import * as vscode from 'vscode';
import fetch from 'cross-fetch';
import { createParser } from 'eventsource-parser';

let controller: AbortController | undefined;
let channel: vscode.OutputChannel;

export function activate(context: vscode.ExtensionContext) {
  channel = vscode.window.createOutputChannel('Agentry SSE');
  context.subscriptions.push(channel);

  context.subscriptions.push(
    vscode.commands.registerCommand('agentry.openPanel', async () => {
      const serverUrl = await vscode.window.showInputBox({
        prompt: 'Agentry server URL',
        value: 'http://localhost:8080',
      });
      if (!serverUrl) {
        return;
      }
      const input = await vscode.window.showInputBox({
        prompt: 'Message to send',
        value: 'hello',
      });
      if (input === undefined) {
        return;
      }
      startStream(serverUrl, input);
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('agentry.stopStream', () => {
      controller?.abort();
      controller = undefined;
      channel.appendLine('\n--- stream stopped ---\n');
    })
  );
}

async function startStream(serverUrl: string, input: string) {
  controller?.abort();
  controller = new AbortController();
  channel.show(true);
  channel.clear();
  channel.appendLine(`Connecting to ${serverUrl}...`);
  let res: any;
  try {
    res = await fetch(`${serverUrl}/invoke`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ input, stream: true }),
      signal: controller.signal,
    });
  } catch (err: any) {
    vscode.window.showErrorMessage(`Failed to connect: ${err.message}`);
    return;
  }

  const parser = createParser((event) => {
    if (event.type === 'event') {
      channel.append(event.data);
    }
  });

  for await (const chunk of res.body as any as AsyncIterable<Uint8Array>) {
    parser.feed(new TextDecoder().decode(chunk));
  }
}

export function deactivate() {
  controller?.abort();
}
