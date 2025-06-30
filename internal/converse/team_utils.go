package converse

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/marcodenic/agentry/internal/tool"
)

// agentNameRegex defines valid agent name pattern: starts with letter, contains letters, numbers, underscores, hyphens
var agentNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

var (
	roleConfigCache  sync.Map
	roleDirOnce      sync.Once
	roleTemplatesDir string
)

// getToolNames extracts tool names from a registry for debugging
func getToolNames(reg tool.Registry) []string {
	var names []string
	for name := range reg {
		names = append(names, name)
	}
	return names
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func findRoleTemplatesDir() string {
	workDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir := workDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		p := filepath.Join(dir, "templates", "roles")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// isValidAgentName checks if an agent name follows the required conventions
func isValidAgentName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	return agentNameRegex.MatchString(name)
}

// mapLegacyToolsToCommands converts old tool names to semantic commands for backward compatibility
func mapLegacyToolsToCommands(legacyTools []string) []string {
	toolMap := map[string][]string{
		"bash":       {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"powershell": {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"cmd":        {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"sh":         {"run", "list", "view", "write", "search", "find", "cwd", "env"},
		"ls":         {"list"},
		"view":       {"view"},
		"read":       {"view"},
		"write":      {"write"},
		"edit":       {"write"},
		"patch":      {"write"},
		"grep":       {"search"},
		"find":       {"find"},
		"fetch":      {}, // fetch is a builtin, not a semantic command
	}

	commandSet := make(map[string]bool)
	for _, tool := range legacyTools {
		if commands, exists := toolMap[tool]; exists {
			for _, cmd := range commands {
				commandSet[cmd] = true
			}
		}
	}

	var result []string
	for cmd := range commandSet {
		result = append(result, cmd)
	}

	return result
}
