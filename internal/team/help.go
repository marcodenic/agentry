package team

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// ProposeCollaboration allows agents to propose working together without inbox messaging
func (t *Team) ProposeCollaboration(ctx context.Context, proposerID, targetAgentID, proposal string) error {
    // Persist proposal metadata in shared store and publish a workspace event
    proposalKey := fmt.Sprintf("proposal_%s_to_%s_%d", proposerID, targetAgentID, time.Now().Unix())
    proposalData := map[string]interface{}{
        "from": proposerID, "to": targetAgentID, "proposal": proposal, "status": "pending", "timestamp": time.Now(),
    }
    t.SetSharedData(proposalKey, proposalData)

    t.PublishWorkspaceEvent(proposerID, "collaboration_proposal", fmt.Sprintf("Proposed collaboration with %s", targetAgentID), map[string]interface{}{
        "target_agent": targetAgentID, "proposal": proposal,
    })

    // Also log a coordination event for observability
    t.LogCoordinationEvent("collab_proposal", proposerID, targetAgentID, proposal, map[string]interface{}{})
    return nil
}

// checkWorkCompleted attempts to detect if an agent completed meaningful work
// even if the response generation timed out
func (t *Team) checkWorkCompleted(agentID, task string) bool {
    taskLower := strings.ToLower(task)
    fileCreationKeywords := []string{
        "create", "write", "generate", "build", "make", "implement",
        "add file", "new file", "script", "code", "project",
        "folder", "directory", "app", "api", "web", "flask", "django",
    }
    hasFileWork := false
    for _, keyword := range fileCreationKeywords {
        if strings.Contains(taskLower, keyword) {
            hasFileWork = true
            break
        }
    }
    if !hasFileWork {
        return false
    }
    workDir, err := os.Getwd()
    if err != nil {
        return false
    }
    recentThreshold := time.Now().Add(-5 * time.Minute)
    hasRecentFiles := false
    filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        if strings.Contains(path, "/.git/") || strings.Contains(path, "/.") {
            return nil
        }
        if info.ModTime().After(recentThreshold) {
            ext := filepath.Ext(path)
            meaningfulExts := map[string]bool{
                ".py": true, ".js": true, ".go": true, ".java": true, ".cpp": true, ".c": true,
                ".html": true, ".css": true, ".json": true, ".yaml": true, ".yml": true,
                ".md": true, ".txt": true, ".sql": true, ".sh": true, ".bat": true,
                ".ts": true, ".jsx": true, ".tsx": true, ".vue": true, ".php": true,
            }
            if meaningfulExts[ext] || info.IsDir() {
                hasRecentFiles = true
                return filepath.SkipDir
            }
        }
        return nil
    })
    return hasRecentFiles
}
