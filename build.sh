#!/usr/bin/env bash
set -euo pipefail

if ! command -v make >/dev/null 2>&1; then
  echo "make is required to use this helper script" >&2
  exit 1
fi

make "${1:-build}"
