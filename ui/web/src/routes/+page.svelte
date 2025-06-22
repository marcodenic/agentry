<script>
  import { onMount } from 'svelte';
  let traces = [];
  let metrics = '';
  let input = '';
  async function send() {
    const resp = await fetch('/invoke', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({agent_id: 'default', input, stream: true})
    });
    const reader = resp.body.getReader();
    const dec = new TextDecoder();
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      const text = dec.decode(value);
      text.trim().split('\n').forEach((line) => {
        if (line.startsWith('data:')) {
          traces.push(JSON.parse(line.slice(5)));
        }
      });
    }
    refresh();
  }
  async function refresh() {
    const res = await fetch('/metrics');
    metrics = await res.text();
  }
  onMount(() => {
    refresh();
    const i = setInterval(refresh, 5000);
    return () => clearInterval(i);
  });
</script>

<input bind:value={input} placeholder="Ask..." />
<button on:click={send}>Send</button>
<h3>Traces</h3>
<pre>{JSON.stringify(traces, null, 2)}</pre>
<h3>Metrics</h3>
<pre>{metrics}</pre>
