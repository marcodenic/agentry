package team

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/marcodenic/agentry/internal/contracts"
    "github.com/marcodenic/agentry/internal/core"
    "github.com/marcodenic/agentry/internal/tokens"
)

// runAgent executes an agent with the given input, similar to converse.runAgent
func runAgent(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
    timer := StartTimer(fmt.Sprintf("runAgent(%s)", name))
    defer timer.Stop()

    // Attach agent name into context for builtins to use sensible defaults
    ctx = context.WithValue(ctx, contracts.AgentNameContextKey, name)
    timer.Checkpoint("context prepared")

    // Minimal bounded context wrapper (idempotent via sentinel)
    contextualInput := buildContextMinimal(ctx, input, name)
    timer.Checkpoint("contextual input built")
    result, err := ag.Run(ctx, contextualInput)
    timer.Checkpoint("agent.Run completed")

    debugPrintf("🏁 runAgent: ag.Run completed for agent %s", name)
    debugPrintf("🏁 runAgent: Result length: %d", len(result))
    debugPrintf("🏁 runAgent: Error: %v", err)
    debugPrintf("🏁 runAgent: Agent %s tokens after: %d", name, func() int {
        if ag.Cost != nil {
            return ag.Cost.TotalTokens()
        }
        return 0
    }())
    debugPrintf("🏁 runAgent: Agent %s context final state: %v", name, ctx.Err())

    return result, err
}

// ---------------- Minimal Context Builder ----------------
const (
    ctxSentinel      = "<!--AGENTRY_CTX_V1-->\n"
    agent0CapTokens  = 1200
    workerCapTokens  = 600
    projectCacheTTL  = 10 * time.Second
)

var (
    projectSummaryCache     string
    projectSummaryCacheLite string
    projectSummaryExpiry    time.Time
    projectSummaryMu        sync.RWMutex
)

func buildContextMinimal(ctx context.Context, task, agentName string) string {
    if strings.HasPrefix(task, ctxSentinel) || os.Getenv("AGENTRY_DISABLE_CONTEXT") == "1" {
        return task
    }
    tier := 1
    if agentName == "agent_0" || agentName == "0" {
        tier = 0
    }
    full, lite := projectSummaries()
    snapshot := lite
    if tier == 0 {
        snapshot = full
    }
    var lines []string
    lines = append(lines, strings.TrimRight(ctxSentinel, "\n"))
    lines = append(lines, snapshot)
    if tier == 1 {
        if refs := extractReferencedFiles(task); len(refs) > 0 {
            lines = append(lines, "Files: "+strings.Join(refs, " "))
        }
    }
    lines = append(lines, "", "TASK:", task)
    assembled := strings.Join(lines, "\n")
    capTokens := workerCapTokens
    if tier == 0 {
        capTokens = agent0CapTokens
    }
    if capEnv := os.Getenv("AGENTRY_CTX_CAP_AGENT0"); tier == 0 && capEnv != "" {
        if v, err := strconv.Atoi(capEnv); err == nil && v > 100 {
            capTokens = v
        }
    }
    if capEnv := os.Getenv("AGENTRY_CTX_CAP_WORKER"); tier == 1 && capEnv != "" {
        if v, err := strconv.Atoi(capEnv); err == nil && v > 100 {
            capTokens = v
        }
    }
    total := tokens.Count(assembled, "gpt-4o-mini")
    if total <= capTokens {
        debugPrintf("CTX agent=%s tier=%d tokens=%d truncated=false\n", agentName, tier, total)
        return assembled
    }
    parts := strings.Split(assembled, "\n")
    idx := 0
    for i, l := range parts {
        if l == "TASK:" {
            idx = i
            break
        }
    }
    firstSnapLine := ""
    for i := 1; i < len(parts) && i < 4; i++ {
        if strings.TrimSpace(parts[i]) != "" {
            firstSnapLine = parts[i]
            break
        }
    }
    header := []string{parts[0]}
    if firstSnapLine != "" {
        header = append(header, firstSnapLine)
    }
    header = append(header, "", "TASK:")
    remaining := strings.Join(parts[idx+1:], "\n")
    allowed := capTokens - tokens.Count(strings.Join(header, "\n"), "gpt-4o-mini") - 10
    if allowed < 50 {
        allowed = 50
    }
    truncated := tokens.Truncate(remaining, allowed, "gpt-4o-mini")
    if tokens.Count(truncated, "gpt-4o-mini") >= allowed && !strings.Contains(truncated, "[truncated]") {
        truncated += "\n...[truncated]"
    }
    final := strings.Join(append(header, truncated), "\n")
    debugPrintf("CTX agent=%s tier=%d tokens=%d truncated=true cap=%d\n", agentName, tier, tokens.Count(final, "gpt-4o-mini"), capTokens)
    return final
}

func projectSummaries() (string, string) {
    now := time.Now()
    projectSummaryMu.RLock()
    if now.Before(projectSummaryExpiry) && projectSummaryCache != "" {
        full, lite := projectSummaryCache, projectSummaryCacheLite
        projectSummaryMu.RUnlock()
        return full, lite
    }
    projectSummaryMu.RUnlock()
    wd, err := os.Getwd()
    if err != nil {
        return "Project: Unknown", "Project: Unknown"
    }
    entries, err := os.ReadDir(wd)
    if err != nil {
        return "Project: Unknown", "Project: Unknown"
    }
    var dirs, configs []string
    projectType := "Unknown"
    for _, e := range entries {
        name := e.Name()
        if e.IsDir() {
            switch name {
            case ".git", ".github", "node_modules", "vendor", ".cache":
            default:
                if len(dirs) < 8 {
                    dirs = append(dirs, name)
                }
            }
            continue
        }
        switch name {
        case "go.mod":
            projectType = "Go"
            configs = append(configs, name)
        case "package.json":
            projectType = "Node"
            configs = append(configs, name)
        case "pyproject.toml", "requirements.txt":
            if projectType == "Unknown" { projectType = "Python" }
            configs = append(configs, name)
        case "Cargo.toml":
            projectType = "Rust"
            configs = append(configs, name)
        case "docker-compose.yml", "Dockerfile", "Makefile":
            configs = append(configs, name)
        }
    }
    if len(dirs) > 5 { dirs = dirs[:5] }
    if len(configs) > 4 { configs = configs[:4] }
    full := fmt.Sprintf("Project: %s; Dirs: %s; Config: %s", projectType, strings.Join(dirs, " "), strings.Join(configs, " "))
    lite := fmt.Sprintf("Project: %s; Dirs: %s", projectType, strings.Join(dirs, " "))
    projectSummaryMu.Lock()
    projectSummaryCache, projectSummaryCacheLite = full, lite
    projectSummaryExpiry = time.Now().Add(projectCacheTTL)
    projectSummaryMu.Unlock()
    return full, lite
}

var allowedFileExt = map[string]struct{}{".go": {}, ".md": {}, ".txt": {}, ".json": {}, ".yaml": {}, ".yml": {}, ".ts": {}, ".js": {}, ".py": {}}

func extractReferencedFiles(task string) []string {
    words := strings.Fields(task)
    seen := make(map[string]struct{})
    var out []string
    for _, w := range words {
        if len(out) >= 5 { break }
        if !strings.Contains(w, ".") { continue }
        w = strings.Trim(w, "`'\"()[]{}<>,")
        if len(w) > 80 { continue }
        ext := filepath.Ext(w)
        if _, ok := allowedFileExt[ext]; !ok { continue }
        if strings.Count(w, "/") > 2 { continue }
        if _, ok := seen[w]; ok { continue }
        if _, err := os.Stat(w); err == nil {
            seen[w] = struct{}{}
            out = append(out, w)
        }
    }
    return out
}

