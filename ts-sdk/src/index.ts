import fetch from "cross-fetch";
import { createParser } from "eventsource-parser";

export interface InvokeOpts {
  agentId?: string;
  stream?: boolean;
  serverUrl?: string;
  onToken?: (tok: string) => void;
}

export async function invoke(
  input: string,
  { agentId, stream, serverUrl = "http://localhost:8080", onToken }: InvokeOpts = {},
): Promise<string> {
  let res;
  try {
    res = await fetch(`${serverUrl}/invoke`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ agent_id: agentId, input, stream }),
    });
  } catch (err: any) {
    const reason = err.message || err.toString();
    throw new Error(
      `❌ Could not connect to Agentry server at ${serverUrl}.\nReason: ${reason}\n\n➡️ To fix: Make sure the server is running (try 'agentry serve') and accessible at this address.`,
    );
  }

  if (!stream) {
    const { output } = await res.json();
    return output;
  }

  let final = "";
  const parser = createParser(evt => {
    if (evt.type === "event") {
      final += evt.data;
      onToken?.(evt.data);
    }
  });
  for await (const chunk of res.body as any as AsyncIterable<Uint8Array>) {
    parser.feed(new TextDecoder().decode(chunk));
  }
  return final;
}

