# Testing & Validation Guide

For the canonical, up-to-date checklist, see [../TEST.md](../TEST.md).

This guide explains how to run the suite locally and what additional scripts exist for manual validation.

---

## Prerequisites

- Go 1.23+
- Optional: `.env.local` with `OPENAI_API_KEY` if you want to hit real models during ad-hoc runs

```bash
cp .env.example .env.local
# export OPENAI_API_KEY=...
```

Run `go mod tidy` after cloning to ensure dependencies are downloaded.

## Core Test Suite

```bash
go test ./...
```

That command executes all unit and integration tests. If you only want the fast subset, pass `-short`.

## Helper Scripts

- `scripts/test_go125_features.sh` – exercises the Go 1.25 feature demos (WaitGroup.Go etc.)
- `scripts/test_debug_logging.sh` – verifies the rolling debug logger end to end
- `scripts/debug-agentry.sh` – wraps `agentry` with verbose logging enabled

## Manual Verification Checklist

1. Start the TUI: `agentry --config .agentry.yaml`
2. Run a direct prompt: `agentry "summarize README"`
3. Refresh model pricing: `agentry refresh-models`
4. Confirm traces: `agentry --trace trace.jsonl "hello"`
5. Inspect logs in `debug/` (or via `scripts/debug-agentry.sh`)

## CI Expectations

GitHub Actions runs `go test ./...` on every PR. Keep the suite fast (<2 minutes on a laptop). If you add new helper scripts, wire them into CI explicitly.

## Troubleshooting

- Ensure `$PATH` includes the directory where `go install` places binaries (`$GOPATH/bin` or `$HOME/go/bin`)
- Delete the Go build cache (`go clean -cache`) if you encounter stale errors
- Set `AGENTRY_DEBUG_LEVEL=trace` when reproducing issues locally

If problems persist, open an issue with the failing command, output, and environment details.
