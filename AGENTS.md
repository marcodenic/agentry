# AGENTS.md

project:
name: agentry
description: A minimal, extensible agentic runtime written in Go, with a TypeScript SDK.

languages:

- go
- typescript

entrypoints:

- cmd/agentry/main.go
- ts-sdk/src/index.ts

test:
run: - go test ./... - cd ts-sdk && npm install && npm test

tools:
go: "1.22"
node: "22"

filetypes:

- "\*.go"
- "\*.ts"
- "\*.yaml"
- "\*.json"
