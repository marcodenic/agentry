# Installation

Prebuilt binaries are available on [GitHub Releases](https://github.com/marcodenic/agentry/releases).

## Homebrew (macOS/Linux)
```bash
brew tap marcodenic/agentry
brew install agentry
```

## Scoop (Windows)
```powershell
scoop bucket add agentry https://github.com/marcodenic/agentry
scoop install agentry
```

## Debian
```bash
wget https://github.com/marcodenic/agentry/releases/download/vX.Y.Z/agentry_X.Y.Z_amd64.deb
sudo dpkg -i agentry_X.Y.Z_amd64.deb
```

## Build from Source

If you have Go installed, you can build and install Agentry directly:

```bash
# Install the latest version (recommended)
go install github.com/marcodenic/agentry/cmd/agentry@latest

# Or install from a local clone
git clone https://github.com/marcodenic/agentry.git
cd agentry
go install ./cmd/agentry
```

### Important Notes for Developers

- **Use `go install`**: This installs the binary to your `$GOPATH/bin` directory where it can be found in your `$PATH`.
- **Avoid `go build` in repo root**: Running `go build` in the repository root creates build artifacts (`agentry`, `agentry.exe`) that should not be committed to git.
- **Check your PATH**: Ensure `$GOPATH/bin` (or `$HOME/go/bin` if `GOPATH` is unset) is in your `$PATH` environment variable.

You can verify the installation with:
```bash
agentry version
```

## Custom Packaging

Homebrew and Scoop formulas are provided under `packaging/`. They are generated using the scripts in `scripts/` and can be reused for custom taps or buckets.

```bash
# create updated formulas after building release binaries
./scripts/package-homebrew.sh <version>
./scripts/package-scoop.ps1 -Version <version>
```

The generated files appear in `packaging/homebrew/` and `packaging/scoop/`. Update the `sha256` fields with the checksums of your release binaries before publishing.
