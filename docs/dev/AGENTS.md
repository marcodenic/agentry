# AGENTS.md

> Keep the root tidy. No scratch files, stray binaries, or experimental scripts.

## Root Hygiene Rules

Root directory must contain only:

- Core project files (`go.mod`, `Makefile`, `README.md`, `PRODUCT.md`, etc.)
- Licensing and contributor docs (`LICENSE`, `CONTRIBUTING.md`, `CODEOWNERS`)
- Generated binary directory (`bin/`) is allowed but should stay gitignored

Everything else belongs in a subdirectory:

| Directory   | Purpose                                            |
| ----------- | -------------------------------------------------- |
| `cmd/`      | CLI entrypoints                                    |
| `internal/` | Runtime implementation (agents, tools, tracing)    |
| `templates/`| Role templates consumed by `.agentry.yaml`         |
| `docs/`     | User and developer documentation                   |
| `scripts/`  | Developer utilities and helper scripts             |
| `tests/`    | Integration and regression suites                  |
| `packaging/`| Release automation (Homebrew, Scoop, Debian)       |

If you need scratch space, use `/tmp` or create a local directory ignored by git.

## Primary Entrypoints

| Name      | Description            | Command                              |
| --------- | ---------------------- | ------------------------------------ |
| `agentry` | Main runtime & CLI     | `agentry`, `agentry "prompt"`        |
| Scripts   | Helper utilities       | `scripts/debug-agentry.sh`, `scripts/test_go125_features.sh` |

## Testing

- Run the full suite with `go test ./...`
- Helper scripts live in `scripts/`

## Documentation Expectations

Every feature change must update the relevant docs:

1. Update [PRODUCT.md](../../PRODUCT.md) if the scope or roadmap changes
2. Adjust README / usage docs for user-facing behaviour
3. Amend this file if directory layout or entrypoints move
4. Keep [docs/testing.md](../testing.md) aligned with new test steps

Documentation updates are part of the definition of done.

## Contribution Guidelines

- New tools go under `internal/tool/` with tests
- Keep prompts/roles in `templates/roles/`
- Coordinate changes through issues or pull requests
- See [CONTRIBUTING.md](../../CONTRIBUTING.md) for review expectations

## Support

Open an issue for questions, bugs, or feature discussions. Keep conversations in public threads for shared context.
