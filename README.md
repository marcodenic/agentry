```
                                 
                                 
    ████▒               ▒████    
      ▒▓███▓▒       ▒▓███▓▒      
        ▒█▒████▓▒▓████▓█▒        
        ▒█   ▓█████▓▒  █▒        
        ▒█▓███▓▓█▓▓███▓█▒        
     ▒▓███▓▒   ▒▓▒   ▒▓███▓▒     
   ▒███▓▓█     ▒▓▒     █▓▓▓██▒   
        ▒█     ▒▓▒     █▒        
        ▒█     ▒▓▒     █▒        
        ▒█     ▒▓▒     █▒        
        ▒█     ▒▓▒     █▒        
        ▒█     ▒▓▒     █▒        
        ▒█     ▒▓▒     █▒        
               ▒▓▒               
                                 
                         v0.2.0  
   █▀█ █▀▀ █▀▀ █▀█ ▀█▀ █▀▄ █ █   
   █▀█ █ █ █▀▀ █ █  █  █▀▄  █    
   ▀ ▀ ▀▀▀ ▀▀▀ ▀ ▀  ▀  ▀ ▀  ▀    
 AGENT  ORCHESTRATION  FRAMEWORK 
```

![Demo](agentry.gif)

Overview
- Minimal Go binary with fast startup and no heavy dependencies
- Built-in TUI for day-to-day coding and debugging
- Pluggable, permission-gated tools defined in `.agentry.yaml`
- Team/delegation helpers so Agent 0 can spawn and coordinate specialists
- Structured tracing plus live token/cost accounting

Install
- Prereq: Go 1.23+
- Install CLI: `go install github.com/marcodenic/agentry/cmd/agentry@latest`

Quick Start
- TUI (default): `./agentry`
- Direct prompt: `./agentry "summarize the README"`
- Show version: `agentry --version`

Configuration
- Project config: `.agentry.yaml` (shipped in the repo root)
- Env vars: copy `.env.example` to `.env.local` and set keys (e.g., `OPENAI_API_KEY`)
- Flags you may care about:
  - `--config PATH`: select config file
  - `--debug`: verbose diagnostics
  - `--allow-tools a,b` / `--deny-tools a,b` / `--disable-tools`
  - `--max-iter N` and `--http-timeout SEC` for runtime tuning

Usage Notes
- TUI launches when no command is provided: just run `agentry`
- You can also pass a direct prompt without a subcommand
- The TUI supports spawning additional agents and shows live token/cost usage

Built-in Tools
- Tools are enabled by listing them in your `.agentry.yaml`
- Core categories: file editing (`view`, `create`, `edit_range`, `search_replace`), search (`ls`, `find`, `grep`, `glob`), shell (`bash`, `sh`, `cmd`, `powershell`), networking (`fetch`, `api`, `download`), delegation (`agent`), diagnostics (`lsp_diagnostics`, `sysinfo`, `ping`)
- Use the allow/deny flags or the config `permissions` block to gate usage for a repository

Tracing & Costs
- Every run can emit structured trace events
- Summaries include input/output tokens and estimated cost per run

Development
- Build: `make build` (outputs `./agentry`)
- Tests: `go test ./...` or `./scripts/test.sh`
- Formatting: CI enforces `gofmt -l` cleanliness

Versioning & Releases
- The internal version constant lives in `internal/version.go`
- Release workflow publishes binaries on tag push like `v0.1.1`

License
- MIT, see `LICENSE`
