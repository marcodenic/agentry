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

## Custom Packaging

Homebrew and Scoop formulas are provided under `packaging/`. They are generated using the scripts in `scripts/` and can be reused for custom taps or buckets.

```bash
# create updated formulas after building release binaries
./scripts/package-homebrew.sh <version>
./scripts/package-scoop.ps1 -Version <version>
```

The generated files appear in `packaging/homebrew/` and `packaging/scoop/`. Update the `sha256` fields with the checksums of your release binaries before publishing.
