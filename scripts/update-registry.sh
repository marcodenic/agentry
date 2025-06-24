#!/bin/bash
set -euo pipefail
FILE="docs/registry/plugins.json"
KEY=${GPG_KEY_ID:-"test@example.com"}

jq -c '.[]' "$FILE" | while read -r line; do
    name=$(echo "$line" | jq -r .name)
    url=$(echo "$line" | jq -r .url)
    sha=$(echo "$line" | jq -r .sha256)
    data="${name}|${url}|${sha}"
    echo -n "$data" >/tmp/data.txt
    sig=$(gpg --batch --yes --local-user "$KEY" --output - --detach-sign /tmp/data.txt | base64 -w0)
    line=$(echo "$line" | jq --arg sig "$sig" '.sig=$sig')
    echo "$line"
done | jq -s '.' > "$FILE"

gpg --armor --output docs/registry/registry.pub --export "$KEY"

