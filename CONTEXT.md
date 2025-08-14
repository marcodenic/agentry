Totally—give each agent the *right* view of the world, not the *same* view. The trick is to make “context” a first-class, pluggable thing with scoring and token budgeting, then compose different packs per agent + task.

Below is a concrete, production-ready way to structure it in your Go framework. You can drop these pieces into your `team` package (or a new `context` subp**Sources:** The context strategies above are drawn from official discussions and docs (GitHub Copilot's description of using current and related files, GitHub's engineering blog on Copilot's context window usage) as well as the design of OpenCode/Crush (loading project files and LSP on session start, using ignore files to limit context, and proposals like a universal AGENT.md for project instructions). These illustrate how **professional AI dev tools manage context** to give the model a comprehensive view of the project within its token limits.

---

# Current Implementation Issues & Solutions

## Problem: Excessive Context Injection (Discovered 2024)

The current `buildContextualInput()` function in `/internal/team/team.go` suffers from:

1. **Token Budget Violations**: Hard-coded context injection leading to >30k tokens, causing Anthropic rate limit errors
2. **Irrelevant Context**: Including entire workspace dumps and coordination history regardless of task relevance
3. **No Dynamic Sizing**: Fixed context blocks that don't adapt to model limits or task requirements
4. **Hardcoded Project Details**: Static project information that doesn't scale across different codebases

## Immediate Fixes Needed

✅ **Root File Tree Context**: 
   - Implemented dynamic project structure detection in `buildRootFileTree()`
   - Replaces hardcoded project details with intelligent file/directory analysis
   - Automatically detects project type (Go, Node.js, Python, Rust) from config files
   - Provides categorized listing: directories, config files, documentation, other files
   - Respects ignore patterns (node_modules, .git, etc.) to avoid clutter

1. **Implement Token Budgeting**: 
   - Calculate available context space: `modelCtx - system - userAsk - guardrails`
   - Allocate budget by provider weights and task relevance
   - Enforce hard limits per context pack

2. **Context Relevance Scoring**:
   - Score files by: semantic similarity, lexical hits, structural affinity, recency, centrality
   - Only include top-k most relevant context packs
   - Use hybrid retrieval (embeddings + search + LSP)

3. **Dynamic Context Packs**:
   - Replace hardcoded workspace dumps with intelligent providers
   - Implement truncation strategies (prefix/suffix for files, outline for large docs)
   - Add provenance metadata (file:line references)

## TODO Tool Integration

Agents need a shared TODO list tool to:
- Track deferred tasks across agent sessions
- Coordinate work between different agent roles
- Maintain context across long-running projects
- Avoid duplicate work and conflicting changes

See `/PRODUCT.md` for detailed TODO tool API specification and implementation plan.ckage) and replace your current hard-coded `buildContextualInput`.

---

# High-level shape

   Each pack contributes a slice of text + metadata (cost, provenance). Examples: *WorkspaceSummary*, *TaskSpec*, *ActiveFile*, *RelatedFiles*, *LSPDefs*, *GitDiff*, *TestFailures*, *RunOutput*, *ChatHistory*, *AgentMemory*, *Rules/AGENT.md*…

2. **Providers** build packs on demand
   Providers know *how* to fetch (LSP, ripgrep/ctags/tree-sitter, embeddings, git). They don’t know which agent will use them.

3. **Profiles** choose which packs an agent gets
   A *Coder* profile might ask for ActiveFile+RelatedFiles+LSPDefs+GitDiff+Tests; a *Planner* profile might prefer WorkspaceSummary+TaskSpec+History.


6. **Loop**
# Core interfaces (drop-in Go)

```go
// context/types.go
    TokensUsed  int
    HardLimit   int            // optional cap
    Name() string
    // Build returns zero or more packs. It should be cheap if nothing relevant.
    AgentName      string
    Task           string
    ModelCtxTokens int
    RecentEvents   []WorkspaceEvent
    RecentEdits    []string // file paths
    Memory         map[string]any
}

type BudgetPlan struct {
    ModelCtxTokens int
    ReservedSystem int
    ReservedUser   int
    PerPack        map[string]int // "ActiveFile"->3000, etc.
}
```

---

# Providers you’ll want (and what they feed)

* **TaskSpecProvider** – the user ask + any agent-role specifics (short).
* **RulesProvider** – `AGENT.md` / `CRUSH.md` / README and conventions (summarized).
* **WorkspaceSummaryProvider** – top-k dirs/files, build/test cmds, detected frameworks.
* **ActiveFileProvider** – prefix/suffix around cursor; configurable window (e.g. 400–1200 lines).
* **RelatedFilesProvider** – choose N files via hybrid search:

  * lexical (ripgrep), structural (ctags/tree-sitter), semantic (embeddings),
  * plus dependency edges (import graph) and recently edited/open buffers.
  * Include *excerpts* around hits, not entire files; show `path:lineStart-lineEnd`.
* **LSPDefsProvider** – `definition`, `references`, `hover` docs for symbols near cursor.
* **GitDiffProvider** – staged & unstaged diffs; last commit msg; WIP markers.
* **TestFailProvider** – last failing tests (names, traces, failing lines).
* **RunOutputProvider** – recent command outputs `go test`, `npm run build`, etc.
* **HistoryProvider** – last K chat turns compacted to bullets.
* **MemoryProvider** – durable breadcrumbs (decisions, TODOs, invariants).
* **IssueProvider** – optional: linked ticket text or PR desc.

Each provider returns 0..N packs (e.g., RelatedFiles might return one pack per file or one merged pack).

---

# Agent profiles (who gets what)

```go
// context/profiles.go
type Profile struct {
    Name      string
    Providers []string
    Weights   map[string]int // importance for budget allocation
}

var Profiles = map[string]Profile{
    "planner": {
        Name: "planner",
        Providers: []string{
            "TaskSpec","Rules","WorkspaceSummary","History","Memory",
        },
        Weights: map[string]int{
            "TaskSpec":3,"Rules":2,"WorkspaceSummary":3,"History":2,"Memory":2,
        },
    },
    "coder": {
        Name: "coder",
        Providers: []string{
            "TaskSpec","ActiveFile","LSPDefs","RelatedFiles","GitDiff","TestFail","RunOutput","Memory",
        },
        Weights: map[string]int{
            "TaskSpec":2,"ActiveFile":5,"LSPDefs":3,"RelatedFiles":4,"GitDiff":2,"TestFail":3,"RunOutput":2,"Memory":1,
        },
    },
    "reviewer": {
        Name: "reviewer",
        Providers: []string{
            "GitDiff","RelatedFiles","Rules","TestFail","RunOutput","History",
        },
        Weights: map[string]int{
            "GitDiff":5,"RelatedFiles":3,"Rules":2,"TestFail":3,"RunOutput":2,"History":1,
        },
    },
}
```

---

# Budgeting & truncation strategy

1. **Compute base budget**: `modelCtx - system - userAsk - guardrails`.
2. **Allocate by weights**; enforce min per critical packs (ActiveFile, TaskSpec).
3. **Pack-level truncators**:

   * Active file: **prefix 200–400 lines + suffix 80–200 lines** around cursor + *file header* (imports, top doc).
   * Related files: top-k (2–5). For each, take **multiple targeted windows** around hits (e.g., 40–120 lines per hit).
   * Rules/README: feed **outline + key sections**, not entire doc. Use headings and bullet summaries.
   * History/Memory: roll-up to bullet points, keep IDs to let agent ask to expand.
4. **Always include provenance** (file paths, line spans) in text so the agent can reference them precisely.

---

# Assembler (replaces your `buildContextualInput`)

```go
// context/assemble.go
func AssemblePrompt(system string, packs []*ContextPack, userTask string) string {
    var b strings.Builder
    b.WriteString(system)
    b.WriteString("\n\n")
    sort.SliceStable(packs, func(i, j int) bool { return packs[i].Priority > packs[j].Priority })
    seen := map[string]struct{}{}

    for _, p := range packs {
        if _, dup := seen[p.Name]; dup && p.Name != "RelatedFiles" { continue }
        seen[p.Name] = struct{}{}
        b.WriteString("## ")
        b.WriteString(p.Name)
        b.WriteString("\n")
        b.WriteString(p.Content)
        b.WriteString("\n\n")
    }
    b.WriteString("## TASK\n")
    b.WriteString(userTask)
    b.WriteString("\n\n")
    b.WriteString("## IMPORTANT\n")
    b.WriteString("- Use create/edit_range/patch/run tools to make REAL changes.\n")
    b.WriteString("- Verify with tests or commands. Cite files/lines you touch.\n")
    return b.String()
}
```

Hook it up in your `runAgent`:

```go
func runAgent(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
    q := contextQueryFromTeam(ctx, name, input) // fill WorkingFile, Cursor, Language, etc.
    prof := Profiles[profileForAgent(name)]     // "coder", "planner", etc.
    packs := BuildPacks(ctx, prof, q)           // use provider registry
    budget := PlanBudget(q.ModelCtxTokens, prof, packs, input)
    packs = ApplyBudget(packs, budget)
    prompt := AssemblePrompt(ag.Prompt, packs, input)
    return ag.Run(ctx, prompt)
}
```

---

# File selection algorithm (for RelatedFiles)

Use a **hybrid score** (0–1) per candidate file:

```
score = 0.45 * semanticSim(task, fileEmb)
      + 0.25 * lexicalHits(tfidf or ripgrep density)
      + 0.15 * structuralAffinity (same package/module/import graph distance)
      + 0.10 * recency (recently edited/open)
      + 0.05 * centrality (degree in call/dep graph)
```

* Build embeddings offline (on save or in background indexing).
* Keep an **import/call graph** (ctags/tree-sitter + LSP).
* Select top-k up to a budget cap; from each file choose 1–3 windows around hits with overlap merging.

---

# LSP integration (minimal but powerful)

* Start language servers based on detected languages (`gopls`, `pyright`, `typescript-language-server`, `rust-analyzer`, etc.).
* Providers:

  * `LSPDefsProvider`: for symbols within ±N lines of cursor, ask `textDocument/definition`, `…/references`, `…/hover`. Pack includes:

    * definition snippet (20–80 lines)
    * type/hover docs (trimmed)
    * reference count + a couple of call sites (short windows)
* Cache per file+version to avoid re-querying every turn.

---

# Memory & history that actually helps

* **Ephemeral (per session)**: summarized chat, last K actions, last K errors, “decisions log”.
* **Durable (project-local)**: small KV under `.agentry/` with:

  * *Invariants* (e.g., “use Cobra CLI”, “errors wrap with `%w`”),
  * *Conventions* (lint rules, log style),
  * *Open threads* (TODOs the agent deferred).
* Providers read both and convert to concise bullets. Keep under \~400–800 tokens total.

---

# Verification loop (self-improvement)

After the coder agent patches files:

1. Run `lint/test/build` (per language pack; infer from Rules/WorkspaceSummary).
2. If failures: `TestFailProvider` and `RunOutputProvider` push fresh packs into the next turn.
3. Gate merges behind the **reviewer** agent:

   * Reviewer profile gets `GitDiff`, `RelatedFiles`, `Rules`, and must emit *blocking comments* with file\:line spans.
4. Planner updates **MemoryProvider** with “what changed” and “why” (short ledger).

---

# Practical truncation defaults (works well)

* Active file: **prefix ≤ 900 lines**, suffix ≤ 200 lines, always include imports + top comments. If file > 2k lines, include an **outline** (function names) plus windows around cursor & current function.
* Related files: **2–5 files**, each **≤ 200–400 lines** in windows.
* LSP: **defs 40–80 lines**, **2–3 call sites @ 20–40 lines** each.
* Git diff: **patch header + hunks**, cap at **300–800 lines**.
* Tests/build output: **last 120–300 lines**, collapse repetitive stack frames.
* Rules/AGENT.md/README: **outline + key sections**, \~300–600 tokens total.
* History/Memory: **≤ 200–400 tokens** combined.

---

# What to change in your code today

* Replace the fixed “WORKSPACE CONTEXT” block with the **Provider → Budget → Assembler** pipeline above.
* Add a `ContextRegistry` on `Team`:

```go
type ContextRegistry map[string]Provider

func (t *Team) RegisterProvider(p Provider) { t.ctxRegistry[p.Name()] = p }
```

* Decide **profile per agent** (you already have `coder`, `writer`, `researcher`; map them to profiles).
* Add a tiny **token estimator** (chars/4 is a decent heuristic) and enforce budgets.
* Persist durable memory into your `memstore` (you already have it—great).

---

# Why this matches (and beats) Copilot/Crush

* Copilot: strong at **active/open files** + a bit of **related code**.
  Your *Coder* profile will include more precise **LSP defs** and **git/test signals** on demand.
* Crush: strong at **project-wide retrieval** + **tool outputs**.
  You’ll match it with Providers for *RunOutput*, *GitDiff*, *TestFail*, and a rules file (**AGENT.md**).

This design keeps prompts lean but rich, scales across languages (LSP-first), and gives each agent the *right* view for its role. It also sets you up for parallelism: multiple agents can build their own context packs concurrently without stepping on each other.

If you want, I can sketch concrete provider structs for Go (LSP via `gopls`, ripgrep search, a simple embeddings index) or wire up the budgeting and truncation helpers.







































Great, I’ll research the full list of contextual signals and inputs used by tools like GitHub Copilot (especially in VS Code), OpenDevin, and Crush to inform LLM-powered agents. I’ll focus on what they extract—such as file context, language server feedback, runtime output, and interaction history—and how much of it is used. I’ll also explore any best practices around tooling, token budgets, or streaming context.

I’ll get back to you with a complete breakdown that you can use to improve your agentic framework’s capabilities.


# **Project Context in VSCode vs. OpenCode/Crush**

Modern AI coding assistants (whether in an IDE like VSCode or in CLI tools like OpenCode/Crush) feed **rich project context** into the model prompt. Below is an **exact breakdown** of what context these tools typically provide, and how much of it is included:

## **VSCode (GitHub Copilot) – Context Sources**

GitHub Copilot (especially with VSCode or Copilot Chat) draws on several layers of context to inform its suggestions:

* **Active File Content:** The primary context is the code in the file you are currently editing. Copilot includes code *before your cursor (prefix)* and *after your cursor (suffix)* in the prompt. Thanks to “Fill-In-the-Middle” (FIM) training, it considers both preceding and following code around the edit point, not just what’s before the cursor. In practice, this often means up to a few hundred lines of the open file are fed into the model (truncated if necessary to fit the model’s context window).

* **Neighboring/Open Files:** Copilot can incorporate *“related files”* beyond the one you’re editing. In fact, recent improvements explicitly pull in content from **other open tabs** in your editor when relevant. For example, if you have a header file or a related module open, Copilot may include portions of those files to provide more context. GitHub noted that *including neighboring open files as context* boosted suggestion relevance by \~5%. Typically, it will select a few of the most relevant open files (or recently edited files) and include their key parts (e.g. function signatures or similar code) in the prompt, as space allows.

* **Relevant Project Code (via Search/Embedding):** Beyond open files, Copilot’s backend algorithms try to find other *“related code”* in your project. Since the model can’t ingest the entire repository (context windows are limited), Copilot uses heuristics and embeddings to pick important snippets. For instance, if you’re calling a function from another file that isn’t open, Copilot might retrieve that function’s definition and include it as extra context. The system searches for the best semantic match (using vector embeddings of code) and adds at least one relevant snippet it finds, rather than nothing. This ensures the model sees crucial definitions or examples from elsewhere in the codebase.

* **User-Provided Context/Instructions:** Copilot also supports custom **instructions and chat context** in VSCode’s Chat. You can explicitly inject files or references using VSCode’s `#` commands (for example, `#codebase` to let it search the whole workspace). When you do this, VSCode will include either the full file content (if it fits) or an **outline/summary** of the file if it’s too large. There’s also a concept of an **AGENT.md / instructions file** (not standard yet) that some tools use to give global project info to AI – e.g. GitHub has an experimental `.github/insights/*.md` for Copilot, and proposals exist to unify this (the **AGENT.md** RFC). In short, if you supply high-level project info (like a README or AGENT.md), Copilot Chat can incorporate that into context, but by default Copilot relies on code context.

* **Chat History & System Prompts:** If using Copilot Chat, the conversation history (your questions and the assistant’s answers) stays in context until the window is full, possibly summarized when it grows too large. Also, behind the scenes Copilot has its own **system prompt** with guidelines. (For example, it was fine-tuned to generate code and may have been given frameworks or style preferences in its system prompt, though those are not user-visible.)

**How much?** In practical terms, older Copilot (based on Codex) had \~2k token context, mainly just the current file. Newer Copilot (GPT-4 based for Copilot Chat) offers much larger context windows (8k or 16k tokens). This means it can include the entire current file (if reasonably sized) plus chunks of a few other files. As GitHub’s team describes: *“not all of a developer’s code can be used as context”*, so the trick is picking the most relevant pieces and ordering them effectively. Copilot’s prompt assembly logic prioritizes the file you’re editing, then open files, then any searched snippets – all fitting within the model’s token limit (e.g. 8,000 tokens). If a file is huge (like thousands of lines), it will truncate or abstract it (as VSCode does by using an outline of functions if needed). The *bottom line*: Copilot gives the model as much of your immediate coding context as possible – primarily the code around your cursor, plus any high-signal related code – without exceeding the context size.

## **OpenCode/Crush (CLI Agents) – Context Sources**

OpenCode (now evolved into **Crush** by Charm) provides an AI coding agent with **even richer project-wide context**. These CLI tools essentially embed the AI in your project directory, so they gather context similarly to how a developer would:

* **Project File Scan (Workspace State):** When you start a session, the tool scans the project structure. For example, **Crush immediately loads your `.gitignore`**, then **finds all source files** in the repo (filtering out ignored files). It identifies the primary language(s) (e.g. seeing many “.go” files means a Go project). This allows it to map out the codebase. Some tools may generate an initial summary of the project – e.g. listing key directories, important files, frameworks detected, etc. (In your code, you attempted this manually with a static list of directories and files). **Crush automates this**, even creating a special context file (like `CRUSH.md`) on init with project-specific notes. For instance, users noticed Crush populating `CHARM.md`/`CRUSH.md` with details like testing conventions (e.g. *“Go tests use PascalCase names”*) gleaned from the codebase. This file (or a standardized **AGENT.md** if present) is given as a *“read me first”* system message to the AI agent. It typically contains project structure info, coding style guidelines, build/test commands, and other high-level context the agent should know.

* **Language Server Protocol (LSP) Integration:** A hallmark of OpenCode/Crush is **deep LSP-enhanced context**. As soon as it identifies the languages in the project, it will spin up the appropriate LSP servers (e.g. `gopls` for Go, `rust-analyzer` for Rust, `pyright` for Python, etc.). The AI agent can query these LSPs on the fly for additional context:

  * *Symbol definitions and usages:* If the AI needs to see where a function or class is defined, it can ask the LSP and retrieve that code or its documentation.
  * *Type information and docs:* The LSP can provide type signatures, comments, or docstrings for functions and structures, which the agent can feed into the prompt when relevant (much like an IDE tooltip, but here it becomes prompt context).
  * *Diagnostics:* LSP error or warning messages (e.g. compiler errors) are accessible, giving the agent awareness of current issues in the code.

  Essentially, **Crush “asks” your development tools for help**, just like a developer would. This real-time context is incorporated when the agent formulates a response or writes code. For example, if you ask “Explain how the auth middleware works,” Crush will use LSP and code search to find the `auth middleware` implementation, pull in those source lines, and only then have the model answer using that context. The answer it gives will even cite file names or line numbers, because it *knows exactly where in the codebase the information came from*.

* **Relevant File Content on Demand:** These agents don’t stuff the entire codebase into the prompt (that’s impossible for large projects); instead, they **retrieve relevant pieces** as needed. They can leverage both simple searching (like `grep` or structural search) and more advanced **embedding-based retrieval** (if implemented) to find which files or sections likely answer the query or are needed for a task. Once identified, those code sections are read and added to the prompt. For instance, OpenCode had a `find` or `search` tool the AI could invoke to locate text in the codebase, and Crush integrates a Model Context Protocol (MCP) for plugging in custom context providers (like documentation or even your database). In practice, if you ask the agent to modify a function, it will `open` that file (bringing its content into context) and then proceed. If you request a new feature, the agent might fetch related modules to ensure consistency. **The amount of code brought in is adjusted to the model’s context size** – e.g. Claude 2 can handle up to 100K tokens, so Crush can afford to include many files at once if using Claude; GPT-4 8k/32k will require more selective loading. The tools aim to include **all relevant code** for the task, and nothing extraneous (irrelevant files are left out, and large files can be partially included or summarized).

* **Project Documentation and Configs:** If your project has docs like README.md, architecture docs, or configuration files (Dockerfiles, build scripts), these tools often incorporate them into context **especially at the start of a session** or when specifically asked. For example, one user of an AI CLI agent (Cascade) found it helpful to feed the README as initial context so the model understands the project’s purpose. Crush by default will load any **special project config** like `.crush.json` or legacy files (.opencode.json, etc.) which may include project-specific instructions. The emerging standard is to have an **AGENT.md** in the repo root with all important info (structure, commands, conventions, etc.) and agents will read it automatically. So, in terms of context, these files serve as a *permanent backdrop* to the conversation – the AI is primed with them so it doesn’t forget fundamentals of your codebase.

* **Persistent Session Memory:** Like ChatGPT, these agent frameworks maintain a conversation history. Every question you ask and every change the agent makes can be remembered. OpenCode had an *“auto-compact”* feature that, when the chat history got too large (95% of model capacity), would automatically summarize older parts of the conversation. That summary stays as context so the model retains important points while freeing up space. This means the agent can **“learn” during the session** – if earlier you explained some requirement or the agent discovered a bug, that knowledge persists. Additionally, these tools log **“coordination events”** (as in your `Team.CoordinationHistory`) and **recent file changes**. Those can be surfaced as context too, usually as a list of the last few actions (e.g. “Agent created file X, modified Y”) to orient the AI. In your code, for example, you append a “Recent Workspace Activity” and “Recent Team Coordination” section to the prompt – Crush does something similar under the hood, showing recent edits and messages among agents so nothing is missed.

* **Tool/Command Outputs:** A unique aspect of agentic AI in the terminal is that the AI can run tools and incorporate their output into context. If the agent runs tests (`go test`, `npm test`) or a shell command (`ls`, `git diff`), the results can be read and fed back into the prompt. This is how Crush can answer questions like *“show me the 5 largest tables”* – it actually runs SQL queries and uses the output as additional context for the model to generate an explanation. So, “context” can include not just static code, but live data from your tools. This capability helps the agent self-improve: it can *verify* its changes by running the code or tests and then adjust based on failures, all within one session’s context.

**How much?** CLI agents can leverage larger context windows more aggressively. With models like Anthropic’s Claude v2 100k or GPT-4 32k, it’s feasible for them to include *dozens of files or long transcripts* when needed. However, they still won’t blindly dump the entire repo. Instead, they intelligently assemble a prompt like:

* High-level project info (from `AGENT.md` or generated rules) – *always included initially* (a few hundred lines at most).
* The specific files/snippets in question – e.g. the function you’re editing or any code a question explicitly references (this could be a few hundred to a few thousand lines, depending on scope).
* A handful of other relevant files found via search or LSP (perhaps brief excerpts or just function signatures if space is tight).
* Recent changes or messages (a short list of bullet points).
* The conversation history (which might be trimmed or summarized as it grows).

All of this is kept within the model’s token limit. For example, if using GPT-4 8k context, the agent might include \~7-8 files worth of content at a time. If using Claude with 100k, it could theoretically include most of a medium-sized codebase in one go, but it will still focus on what’s relevant to the task (and might rely on embeddings to choose which files to load). Additionally, **Crush respects `.crushignore`** (and your `.gitignore`) to exclude huge or irrelevant files from ever being considered – e.g. you can ignore big data files or auto-generated code so they don’t clog the context. This ensures the **token budget is spent only on useful context**.

## **Summary – Achieving Copilot-Level Quality**

To get **Copilot/GPT-4 quality output** from your own agent framework, you’re right that you need to improve its context input. In summary, the best practice is to provide **multi-faceted project context** to the AI, including:

* **Immediate code context** (the code being worked on + surrounding scope).
* **Broad project knowledge** (overview of structure, key conventions, build/test info – via a file like AGENT.md or initial prompt rules).
* **On-demand relevant code retrieval** (use LSP or search to fetch definitions, usages, or related modules whenever the task calls for them).
* **Cross-file awareness** (if a change in one file affects another, feed both files to the model).
* **Memory of recent events** (recent edits, commands run, errors encountered).
* **Leveraging larger models** (use 16k, 32k, or 100k-token models so you can include more context without truncation).

Both VSCode’s Copilot and OpenCode/Crush do some or all of the above. Copilot started with just the open file, but now even it considers multiple files and intelligent snippets. Crush, being purpose-built for projects, goes further by integrating with your development tools and repository content (even running code or queries) to supply *rich context*. By assembling a prompt with the **right pieces of your project at the right time**, these tools enable the AI to act like a knowledgeable pair-programmer who “knows” your entire codebase. Adopting similar strategies (LSP integration, context files, intelligent retrieval) in your agent framework will significantly boost its output quality, moving it closer to human-level assistance and allowing it to **self-improve** by always learning from the latest project state.

**Sources:** The context strategies above are drawn from official discussions and docs (GitHub Copilot’s description of using current and related files, GitHub’s engineering blog on Copilot’s context window usage) as well as the design of OpenCode/Crush (loading project files and LSP on session start, using ignore files to limit context, and proposals like a universal AGENT.md for project instructions). These illustrate how **professional AI dev tools manage context** to give the model a comprehensive view of the project within its token limits.
