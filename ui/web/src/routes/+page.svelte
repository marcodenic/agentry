<script>
  import { onMount } from 'svelte';
  import { Chart } from 'chart.js/auto';
  let traces = [];
  let otel = [];
  let metrics = '';
  let input = '';
  let agents = [];
  let chart;
  let reqChart;
  let usageChart;
  let healthChart;
  let labels = [];
  let tokens = [];
  let requests = {};
  let tokenUsage = {};
  let health = {};

  async function loadAgents() {
    const res = await fetch('/agents');
    if (res.ok) agents = await res.json();
  }
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
    parseMetrics(metrics);
    const usageRes = await fetch('/token_usage');
    if (usageRes.ok) updateUsage(await usageRes.json());
    const healthRes = await fetch('/agent_health');
    if (healthRes.ok) updateHealth(await healthRes.json());
    const tr = await fetch('/traces');
    if (tr.ok) otel = await tr.json();
  }
  function parseMetrics(text) {
    const lines = text.split('\n');
    const tokLine = lines.find((l) => l.startsWith('agentry_tokens_total'));
    if (!tokLine) return;
    const parts = tokLine.split(' ');
    const v = parseFloat(parts[1]);
    tokens.push(v);
    labels.push(new Date().toLocaleTimeString());
    if (tokens.length > 20) { tokens.shift(); labels.shift(); }
    if (!chart) {
      const ctx = document.getElementById('tokChart');
      if (!ctx) return;
      chart = new Chart(ctx, { type: 'line', data: { labels, datasets:[{label:'tokens', data: tokens}] }});
    } else {
      chart.data.labels = labels;
      chart.data.datasets[0].data = tokens;
      chart.update();
    }

    requests = {};
    lines.forEach((l) => {
      if (l.startsWith('agentry_http_requests_total')) {
        const m = l.match(/path="([^"]+)".* (\d+(?:\.\d+)?)/);
        if (m) requests[m[1]] = parseFloat(m[2]);
      }
    });
    const rlabels = Object.keys(requests);
    const rdata = Object.values(requests);
    if (!reqChart) {
      const ctx = document.getElementById('reqChart');
      if (ctx) reqChart = new Chart(ctx, {type:'bar', data:{labels:rlabels, datasets:[{label:'requests', data:rdata}]}});
    } else {
      reqChart.data.labels = rlabels;
      reqChart.data.datasets[0].data = rdata;
      reqChart.update();
    }
  }

  function updateUsage(data) {
    tokenUsage = data;
    const labels = Object.keys(data);
    const vals = Object.values(data);
    if (!usageChart) {
      const ctx = document.getElementById('usageChart');
      if (ctx) usageChart = new Chart(ctx, {type:'bar', data:{labels,datasets:[{label:'tokens', data:vals}]}});
    } else {
      usageChart.data.labels = labels;
      usageChart.data.datasets[0].data = vals;
      usageChart.update();
    }
  }

  function updateHealth(data) {
    health = data;
    const labels = Object.keys(data);
    const vals = Object.values(data);
    if (!healthChart) {
      const ctx = document.getElementById('healthChart');
      if (ctx) healthChart = new Chart(ctx, {type:'bar', data:{labels,datasets:[{label:'up', data:vals}]}, options:{scales:{y:{min:0,max:1}}}});
    } else {
      healthChart.data.labels = labels;
      healthChart.data.datasets[0].data = vals;
      healthChart.update();
    }
  }
  onMount(() => {
    loadAgents();
    refresh();
    const i = setInterval(refresh, 5000);
    return () => clearInterval(i);
  });
</script>

<input bind:value={input} placeholder="Ask..." />
<button on:click={send}>Send</button>
<h3>Running Agents</h3>
<ul>
  {#each agents as a}
    <li>{a}</li>
  {/each}
</ul>
<h3>Traces</h3>
<pre>{JSON.stringify(traces, null, 2)}</pre>
<h3>OTLP</h3>
<pre>{JSON.stringify(otel, null, 2)}</pre>
<h3>Metrics</h3>
<canvas id="tokChart"></canvas>
<canvas id="reqChart"></canvas>
<canvas id="usageChart"></canvas>
<canvas id="healthChart"></canvas>
<pre>{metrics}</pre>
