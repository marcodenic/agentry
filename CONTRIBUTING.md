# Contributing to Agentry

Thank you for your interest in contributing to Agentry! This document provides guidelines for contributors.

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Getting Started

1. Fork the repository
2. Clone your fork:

   ```bash
   git clone https://github.com/yourusername/agentry.git
   cd agentry
   ```

3. Install dependencies:

   ```bash
   go mod download
   ```

4. Install the development version:

   ```bash
   go install ./cmd/agentry
   ```

5. Verify the installation:
   ```bash
   agentry version
   ```

### Building and Testing

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test files
go test ./tests/
```

#### Building Components

```bash
# Build all components
make build

# Build specific components
go build ./cmd/agentry
go build ./cmd/agent-hub
go build ./cmd/agent-node
```

⚠️ **Important**: Do not run `go build` in the repository root directory as this creates build artifacts that should not be committed. Use `go install` for installing binaries or `go build ./cmd/...` for building specific components.

### Code Style

- Follow standard Go formatting with `gofmt`
- Run `go vet` to check for common issues
- Use meaningful variable and function names
- Add comments for exported functions and types

### Submitting Changes

1. Create a feature branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Ensure all tests pass
5. Commit with a clear message
6. Push to your fork
7. Create a pull request

### Git Workflow

- Keep commits atomic and focused
- Write clear commit messages
- Rebase feature branches before submitting PRs
- Use `git status` to verify you're not committing build artifacts

### Common Issues

#### Build Artifacts in Git

If you accidentally run `go build` in the root and have `agentry` or `agentry.exe` files:

```bash
# Remove the artifacts
rm -f agentry agentry.exe

# They should already be in .gitignore, but verify:
git status
```

#### Testing New Tools

When adding new built-in tools:

1. Add the tool to `internal/tool/manifest.go`
2. Add comprehensive tests in `tests/`
3. Update documentation in `docs/api.md` and `docs/usage.md`
4. Add examples to configuration files if appropriate

### Documentation

- Update relevant documentation for any user-facing changes
- Add examples for new features
- Keep README.md and docs/ in sync

## Questions?

If you have questions about contributing, please open an issue or start a discussion on GitHub.
