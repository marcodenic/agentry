# Agentry Configuration Guide

This document explains the purpose of each configuration file in the Agentry project and provides guidelines for maintaining them.

## Architecture Overview

Agentry now uses a **simplified architecture** focused on:
- **Context-Lite Prompt Construction**: Direct message building without complex context providers
- **Direct Agent-to-Agent Delegation**: Using the `agent` tool for task delegation
- **TODO-Driven Workflow**: Built-in TODO management and tracking
- **JSON Output Validation**: Automatic validation of tool arguments and responses
- **Standard Operating Procedures (SOPs)**: Runtime guidance for agent behavior
- **TUI TODO Board**: Visual TODO management in the terminal interface

## Main Configuration Files

### `.agentry.yaml` (Root)
- Purpose: Primary configuration for production use
- Model: Uses `gpt-5` for Agent 0 (system/orchestrator)
- Tools: Comprehensive built-in tools including file operations, web tools, and shell commands
- Features: Direct agent delegation, TODO management, LSP diagnostics
- Usage: Used by default when running `agentry` command

### Tool Landscape
The current tool set focuses on:
- **File Operations**: `view`, `create`, `edit_range`, `search_replace`, `ls`, `find`, `grep`
- **Web Tools**: `web_search`, `read_webpage`, `fetch`, `api`, `download`  
- **Shell Tools**: Platform-specific `bash`, `sh`, `powershell`, `cmd`
- **Agent Coordination**: `agent` for direct delegation
- **Development Tools**: `lsp_diagnostics`, `patch`, `project_tree`
- **TODO Management**: Built-in TODO CRUD operations
- **System Tools**: `sysinfo`, `ping`, `echo`

**Removed Legacy Features**:
- ❌ Inbox messaging (`send_message`, `inbox_read`, `inbox_clear`, `request_help`)
- ❌ Parallel agents tool (`parallel_agents`)
- ❌ Complex context pipeline (replaced with Context-Lite)
- ❌ Auto-related-files detection

### `examples/.agentry.yaml`
- **Purpose**: Example configuration for users to copy and modify
- **Model**: Includes both mock and OpenAI models for flexibility
- **Tools**: Standard set of tools with good documentation
- **Usage**: Reference for new users setting up Agentry

## Testing Configurations (`/config`)

### `test-config.yaml`
- **Purpose**: General testing with full tool set
- **Model**: `gpt-4.1-nano` (ultra cost-effective for testing)
- **Tools**: Comprehensive tool set for testing all functionality
- **Usage**: Used by integration tests and general testing

### `test-delegation-config.yaml`
- **Purpose**: Specifically tests agent delegation functionality
- **Model**: Mock model (no API costs)
- **Tools**: Minimal set focusing on delegation (`agent`, `echo`, `ping`)
- **Usage**: Unit tests for agent delegation

### `persistent-config.yaml`
- **Purpose**: Tests legacy persistent agent functionality
- **Model**: Claude Opus (for advanced testing)
- **Tools**: Core tools plus agent delegation
- **Features**: Legacy persistent-agent knobs (port settings now removed in runtime)
- **Usage**: Tests for persistent agent sessions

### `smart-config.yaml`
- **Purpose**: Advanced configuration with enhanced tools
- **Model**: Multiple models (OpenAI + Anthropic)
- **Tools**: Enhanced tool set with advanced file operations
- **Usage**: Testing advanced features and capabilities

### Session Testing Configs
- **`session-test-config.yaml`**: Tests persistent agents with GPT-4
- **`simple-session-config.yaml`**: Minimal session testing with no tools

## Specialized Test Configurations (`/tests`)

### `tests/bash-tool/`
- **`bash-test-config.yaml`**: Tests shell tools across platforms
- **`windows-test-config.yaml`**: Windows-specific tool testing
- **`direct-test.yaml`**: Direct tool execution testing

## Template Configurations (`/templates`)

### Role Templates (`/templates/roles/`)
- **`agent_0.yaml`**: System/orchestrator agent configuration
- **`coder.yaml`**: Coding specialist agent
- **`tester.yaml`**: Testing specialist agent
- **`researcher.yaml`**: Research specialist agent
- **Other roles**: Various specialized agent roles

### Team Templates
<!-- Removed old team YAML examples (dev_team.yaml, docs_team.yaml, website_team.yaml) during cleanup; refer to role-based includes in smart-config instead. -->

## Configuration Standards

### Model Selection
- Agent 0: Use `gpt-5` for the system/orchestrator role
- Test configs: Use `mock` or low-cost models for CI
- Specialist agents: Choose per-role models in templates

### Format Standards
- Use `models:` array format (not singular `model:`)
- Consistent indentation (2 spaces)
- Clear comments explaining purpose
- Environment variables for API keys

### Tool Configuration
- Always include `agent` tool for delegation
- Include appropriate tools for the use case
- Add `lsp_diagnostics` when working with Go, TypeScript, Python, Rust, or JavaScript projects; it auto-discovers files and runs available tools
- Document tool descriptions clearly

## Maintenance Guidelines

1. **Remove Legacy**: No `routes` or `if_contains` configurations
2. **Consistency**: Use same model naming and format across files
3. **Documentation**: Each config should have clear comments
4. **Testing**: Verify configs work with current codebase
5. **Cleanup**: Remove duplicate or unused configurations

## Configuration Hierarchy

1. **Command line flags** (highest priority)
2. **Environment variables**
3. **Local `.agentry.yaml`** (project-specific)
4. **Global config** (user home directory)
5. **Default values** (lowest priority)

## Common Issues to Avoid

- ❌ Using old `model:` format instead of `models:`
- ❌ Including `routes` or `if_contains` configurations
- ❌ Inconsistent model names across configs
- ❌ Missing required tools like `agent` for delegation
- ❌ Hardcoded API keys (use environment variables)
- ❌ Duplicate configurations without clear purpose
