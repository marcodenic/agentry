# Agentry Cloud Architecture

This document outlines the planned cloud deployment model for Agentry.

## Overview

Agentry exposes a minimal HTTP API backed by the Go runtime. Clients (CLI, SDK, or other services) send requests to the API which orchestrates tools and models.

```
[Client] -> [HTTP API] -> [Agent Runtime] -> [Tools/LLM]
                      -> [Memory]
```

```mermaid
flowchart LR
    subgraph Frontend
        CLI
        SDK
    end
    CLI --> API
    SDK --> API
    API((HTTP API)) --> Runtime
    Runtime --> Tools
    Runtime --> LLM((LLM))
    Runtime --> Memory
```

## Repository Layout

```
.
├── cmd/           # CLI entrypoints
├── internal/      # core runtime packages
├── ts-sdk/        # TypeScript SDK
├── design/        # architecture notes
├── examples/      # sample configs and data
├── tests/         # Go test suites
└── README-cloud.md
```

This layout may evolve as the distributed scheduler and sandbox features mature.

## Building the Dashboard

The web dashboard lives under `ui/web` and uses SvelteKit.

```bash
cd ui/web
npm install
npm run build
```

The build output in `ui/web/dist` is embedded into the hub and served at `/` when running `agentry serve --metrics`.

## Using the Dashboard

Start the server with metrics enabled and then open `http://localhost:8080` in a browser:

```bash
agentry serve --config examples/.agentry.yaml --metrics
```

The dashboard shows:

- **Running Agents** – list of available agent IDs from `/agents`.
- **Traces** – live SSE events plus OTLP spans fetched from `/traces`.
- **Metrics Graphs** – Prometheus counters visualised in real time from `/metrics`.

Data refreshes every few seconds to provide a near real‑time view of the system.
