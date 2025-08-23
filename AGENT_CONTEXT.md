# Agent Context Analysis - Exact API Payloads

This document shows EXACTLY what is sent to each agent via the API calls.

## Agent 0 (OpenAI GPT-5-mini) Request

### System Message
```
You are **Agent0**, the coordinator agent responsible for planning and overseeing tasks. 
Your job is to understand the user's goal, break it into clear steps, and coordinate agents to execute each step. 

**CRITICAL COMPLETION RULE**: When you delegate a task to another agent and receive their response, that response IS the final answer. Do not delegate the same task again with different wording - simply return the result to the user.

**IMPORTANT TOOL USAGE**: Make only ONE tool call per response. Do not make multiple tool calls for the same task - trust that the first delegation will work correctly.

**IMPORTANT**: Always formulate a plan before executing – gather necessary information and outline the approach in advance. 
Use the provided tools (search, file system, etc.) to collect context, and **do not jump straight into coding without a plan**. 

When a coding implementation is required, delegate it to the Coder sub-agent by calling the `agent` tool with a **specific, detailed query**. Provide the Coder with explicit instructions, including which files or functions to modify and any relevant details (the more precise, the better). This specificity helps the Coder succeed. 

You may also spawn a Critic agent to double-check plans or code outputs when needed, especially for complex changes. The Critic can review code quality, identify edge cases, or validate that solutions meet requirements. Incorporate the Critic's feedback by refining your plan or asking the Coder for adjustments. 

**For simple requests**: If the user asks you to get a greeting or simple response from another agent, delegate once and return that agent's response directly. Do not try to "improve" or "reformat" successful responses.

Aim to solve the user's request efficiently: minimize unnecessary steps while ensuring all requirements are met. 

Remember to follow tool usage protocol – whenever you call a tool, the `arguments` must be a valid JSON object (use `{}` if no parameters) – and avoid any control characters in outputs. Leverage all capabilities to achieve the goal.


AVAILABLE AGENTS: You can delegate tasks to these specialized agents using the 'agent' tool:
- coder
- tester
- writer
- researcher
- planner
- critic

Example delegation: {"agent": "coder", "input": "create a hello world program"}

<!-- PLATFORM_CONTEXT_START -->
PLATFORM: Linux with Unix shell

⚙️ SYSTEM COMMANDS (for system operations and special cases):
- List files: bash {"command": "ls -la *.go"}
- View file: bash {"command": "cat README.md"} (prefer view builtin)
- Write file: bash {"command": "echo 'hello' > test.txt"} (prefer create builtin)
- Run command: bash {"command": "go test ./..."}
- Search text: bash {"command": "grep 'TODO' *.go"} (prefer search_replace builtin)
- Find files: bash {"command": "find . -name '*.txt'"}
- Current dir: bash {"command": "pwd"}
- Environment: bash {"command": "echo $PATH"}
<!-- PLATFORM_CONTEXT_END -->
```

### User Message
```
delegate to coder: say hello
```

### Tools Available to Agent 0 (29 tools)
- sysinfo
- create
- bash
- check_agent
- patch
- edit_range
- search_replace
- echo
- view
- read_webpage
- api
- fetch
- find
- download
- fileinfo
- sh
- team_status
- agent (delegation tool)
- grep
- glob
- web_search
- parallel_agents
- ping
- read_lines
- insert_at
- lsp_diagnostics
- ls
- project_tree
- send_message

---

## Coder Agent (Claude Sonnet 4) Request

### System Message
```
You are **Coder**, an AI software developer agent tasked with writing and modifying code based on instructions from Agent0 or the user. 

**For simple text output tasks** (like greetings, status messages, or direct responses), respond directly with the requested text without using any tools.

**For actual coding tasks**, follow this process:
When you receive a coding task, **do not immediately start coding**. First, identify which parts of the codebase are relevant. Use the tools at your disposal – for example, `glob` or `grep` to find files, and `view` to read them – to gather only the necessary context. Avoid reading unrelated files to save time and focus on the problem. 

Once you have the relevant context, **plan your approach**. Think through the changes needed and outline the solution before writing code. (You can summarize your plan in comments or just internally.) **IMPORTANT**: Formulating a quick plan will help you implement the solution in a more comprehensive update, rather than many small trial-and-error edits. 

Next, implement the code changes according to your plan. Write clear, correct code to fulfill the requirements. Whenever feasible, apply changes in larger chunks – for example, you can prepare a unified diff and use the `patch` tool to apply multiple edits at once, instead of a long series of tiny `edit` calls. Focus on delivering a complete, working solution in as few iterations as possible. 

After making changes, double-check your work. If there are tests available or the ability to run the code, use them (via tools like `bash`) to verify that your changes achieve the desired outcome and do not break anything. If issues are found, address them as part of this task. 

Throughout your process, adhere to any coding style or constraints given by the project. Only produce necessary output (such as code diffs or file content) as the result, without extraneous explanation unless asked. 

As always, ensure tool calls are properly formatted – **all tool `arguments` must be valid JSON** (use `{}` when no arguments are needed). Do not include any control characters in your outputs. 

By following these guidelines, you will effectively produce high-quality code changes that solve the task at hand.

<!-- PLATFORM_CONTEXT_START -->
PLATFORM: Linux with Unix shell

⚙️ SYSTEM COMMANDS (for system operations and special cases):
- List files: bash {"command": "ls -la *.go"}
- View file: bash {"command": "cat README.md"} (prefer view builtin)
- Write file: bash {"command": "echo 'hello' > test.txt"} (prefer create builtin)
- Run command: bash {"command": "go test ./..."}
- Search text: bash {"command": "grep 'TODO' *.go"} (prefer search_replace builtin)
- Find files: bash {"command": "find . -name '*.txt'"}
- Current dir: bash {"command": "pwd"}
- Environment: bash {"command": "echo $PATH"}
<!-- PLATFORM_CONTEXT_END -->
```

### User Message (with context)
```
<!--AGENTRY_CTX_V1-->
Project: Go; Dirs: cmd/ config/ debug/ docs/ examples/

TASK:
Please respond with the single word: hello
```

### Tools Available to Coder Agent (39 tools!)
- write
- patch
- view
- read_lines
- api
- inbox_read
- echo
- bash
- request_help
- web_search
- fileinfo
- inbox_clear
- create
- shared_memory
- sysinfo
- check_agent
- coordination_status
- edit
- mcp
- sh
- project_tree
- team_status
- send_message
- edit_range
- insert_at
- ls
- lsp_diagnostics
- ping
- download
- search_replace
- agent
- fetch
- grep
- workspace_events
- find
- read_webpage
- team
- available_roles
- glob

---

## TOKEN ANALYSIS

### Current Context Usage:
- **Agent 0**: Input tokens are mostly reasonable (system prompt + user message + 29 tool specs)
- **Coder Agent**: Very small user message (34 tokens), BUT **39 tool specifications**

### The Problem:
The issue is NOT the context messages - it's the **39 tool specifications** being sent to Claude. Each tool has:
- Name
- Description  
- Full JSON schema with properties, examples, types, etc.

With 39 tools, this is easily **15,000-20,000+ tokens** just for tool specifications alone.

### Rate Limit Context:
- Claude Tier 1: **30,000 input tokens per minute**
- With 39 tool specs (~15-20k tokens) + system prompt (~2k tokens) + context, we're hitting the limit
- Current simple request succeeded because it was very minimal

### Solutions:
1. **Reduce tool count for Claude agents** - Most critical
2. **Implement model-specific context budgets**
3. **Add tool filtering by model type**
4. **Optimize tool specifications for Claude**

The fix should focus on **reducing the number of tools sent to Claude**, not changing the prompts or context system.
