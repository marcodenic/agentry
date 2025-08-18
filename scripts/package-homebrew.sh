#!/usr/bin/env bash
set -euo pipefail

VERSION="$1"
DARWIN_BINARY="dist/agentry-darwin-amd64"
LINUX_BINARY="dist/agentry-linux-amd64"

DARWIN_SHA=$(shasum -a 256 "$DARWIN_BINARY" | awk '{print $1}')
LINUX_SHA=$(shasum -a 256 "$LINUX_BINARY" | awk '{print $1}')

FORMULA_DIR="dist/homebrew"
mkdir -p "$FORMULA_DIR"
FORMULA="$FORMULA_DIR/agentry.rb"

cat > "$FORMULA" <<EOF2
class Agentry < Formula
  desc "Minimal, performant AI-Agent runtime"
  homepage "https://github.com/marcodenic/agentry"
  version "$VERSION"

  on_macos do
    url "https://github.com/marcodenic/agentry/releases/download/v#{version}/agentry-darwin-amd64"
    sha256 "$DARWIN_SHA"
  end

  on_linux do
    url "https://github.com/marcodenic/agentry/releases/download/v#{version}/agentry-linux-amd64"
    sha256 "$LINUX_SHA"
  end

  def install
    bin.install "agentry"
  end
end
EOF2

printf "Homebrew formula written to %s\n" "$FORMULA"
