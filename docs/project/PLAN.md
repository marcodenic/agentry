# Agentry Project Plan

_Last updated: September 2025_

Agentry is now focused on a lean, local-first runtime. The near-term plan concentrates on hardening the core loop, tightening the developer experience, and improving test/documentation coverage.

## 1. Core Runtime

| Task | Notes |
| ---- | ----- |
| Context-Lite prompt compiler | Finalise the simplified system prompt builder and add golden tests |
| Error-handling polish | Exercise consecutive tool error limits and improve messaging |
| Cost manager consistency | Normalise `provider/model` naming and ensure refresh cache is honoured everywhere |

## 2. Tooling & TUI

| Task | Notes |
| ---- | ----- |
| TODO board refinements | Surface status changes and filters inside the TUI |
| Tool UX | Review tool help output and ensure common flags/fields are documented |
| Debug logging ergonomics | Ship slimmer presets for `scripts/debug-agentry.sh` (trace vs. info modes) |

## 3. Test & Docs

| Task | Notes |
| ---- | ----- |
| Internal/core charters | Continue converting charters in `REFACTOR.md` into real tests |
| Configuration docs | Keep `.agentry.yaml` examples aligned with the trimmed feature set |
| Contributor flow | Refresh CONTRIBUTING.md and developer doc pointers |

## Recently Completed

- Removed legacy persistent mode, NATS queueing, and Kubernetes deployment artefacts
- Simplified CLI to `tui`, direct prompt, and `refresh-models`
- Relocated stray scripts/documentation into dedicated directories
- Added regression tests for streaming aggregator and tool error paths

## Out of Scope / Archived

- Cloud orchestration, autoscalers, and long-lived session daemons
- Parallel tool runners and inbox-style messaging

Update this document when priorities shift so the team stays aligned.
