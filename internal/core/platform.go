package core

import (
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/tool"
)

// GetPlatformContext returns OS-specific guidance for agents with tiered tool hierarchy
func GetPlatformContext(allowedCommands []string, allowedBuiltins []string) string {
	var platformInfo string

	switch runtime.GOOS {
	case "windows":
		platformInfo = "PLATFORM: Windows with PowerShell"
	case "darwin":
		platformInfo = "PLATFORM: macOS with Unix shell"
	case "linux":
		platformInfo = "PLATFORM: Linux with Unix shell"
	default:
		platformInfo = "PLATFORM: Unknown OS"
	}

	result := platformInfo + "\n\n"

	// Tier 1: Enterprise-grade builtin tools (PREFERRED)
	if len(allowedBuiltins) > 0 {
		fileOpsTools := getFileOperationTools(allowedBuiltins)
		webTools := getWebTools(allowedBuiltins)
		otherTools := getOtherBuiltinTools(allowedBuiltins)

		if len(fileOpsTools) > 0 || len(webTools) > 0 || len(otherTools) > 0 {
			result += "🎯 PREFERRED TOOLS (use these first):\n"

			if len(fileOpsTools) > 0 {
				result += "\n📁 File Operations (enterprise-grade, atomic, cross-platform):\n"
				for _, tool := range fileOpsTools {
					result += "- " + tool + ": " + getBuiltinDescription(tool) + "\n"
				}
			}

			if len(webTools) > 0 {
				result += "\n🌐 Web & Network Operations:\n"
				for _, tool := range webTools {
					result += "- " + tool + ": " + getBuiltinDescription(tool) + "\n"
				}
			}

			if len(otherTools) > 0 {
				result += "\n🔧 Other Tools:\n"
				for _, tool := range otherTools {
					result += "- " + tool + ": " + getBuiltinDescription(tool) + "\n"
				}
			}

			result += "\n"
		}
	}

	// Tier 2: Shell commands (FALLBACK)
	if len(allowedCommands) > 0 {
		result += "⚙️ SYSTEM COMMANDS (for system operations and special cases):\n"
		commandMap := getCommandMap()
		for _, cmd := range allowedCommands {
			if cmdExample, exists := commandMap[cmd]; exists {
				result += "- " + cmdExample + "\n"
			}
		}
	}

	return result
}

// getFileOperationTools returns file operation builtins
func getFileOperationTools(allowedBuiltins []string) []string {
	fileOps := []string{"read_lines", "edit_range", "insert_at", "search_replace", "fileinfo", "view", "create"}
	var result []string
	for _, tool := range fileOps {
		if contains(allowedBuiltins, tool) {
			result = append(result, tool)
		}
	}
	return result
}

// getWebTools returns web-related builtins
func getWebTools(allowedBuiltins []string) []string {
	webOps := []string{"web_search", "read_webpage", "api", "download", "fetch"}
	var result []string
	for _, tool := range webOps {
		if contains(allowedBuiltins, tool) {
			result = append(result, tool)
		}
	}
	return result
}

// getOtherBuiltinTools returns other builtin tools
func getOtherBuiltinTools(allowedBuiltins []string) []string {
	fileOps := []string{"read_lines", "edit_range", "insert_at", "search_replace", "fileinfo", "view", "create"}
	webOps := []string{"web_search", "read_webpage", "api", "download", "fetch"}
	var result []string

	for _, tool := range allowedBuiltins {
		if !contains(fileOps, tool) && !contains(webOps, tool) {
			result = append(result, tool)
		}
	}
	return result
}

// getBuiltinDescription returns a description for builtin tools
func getBuiltinDescription(tool string) string {
	descriptions := map[string]string{
		"read_lines":     "Read specific lines with line-precise access",
		"edit_range":     "Replace line ranges atomically",
		"insert_at":      "Insert lines at specific positions",
		"search_replace": "Advanced search/replace with regex",
		"fileinfo":       "Comprehensive file analysis",
		"view":           "Enhanced file viewing with line numbers",
		"create":         "Create files with overwrite protection",
		"web_search":     "Search the web for information",
		"read_webpage":   "Extract content from web pages",
		"api":            "Make HTTP/REST API calls",
		"download":       "Download files from URLs",
		"fetch":          "Download content from URLs",
		"agent":          "Delegate tasks to specialized agents",
		"patch":          "Apply unified diff patches",
		"echo":           "Repeat/output text",
		"ping":           "Test network connectivity",
		"branch-tidy":    "Clean up Git branches",
		"mcp":            "Connect to MCP servers",
		"sysinfo":        "Get system information and hardware specs",
	}
	if desc, exists := descriptions[tool]; exists {
		return desc
	}
	return "Advanced tool"
}

// getCommandMap returns OS-specific command examples
func getCommandMap() map[string]string {
	switch runtime.GOOS {
	case "windows":
		return map[string]string{
			"list":   `List files: powershell {"command": "Get-ChildItem -Name '*.go'"}`,
			"view":   `View file: powershell {"command": "Get-Content README.md"} (prefer view builtin)`,
			"write":  `Write file: powershell {"command": "Set-Content -Path test.txt -Value 'hello'"} (prefer create builtin)`,
			"run":    `Run command: powershell {"command": "go test ./..."}`,
			"search": `Search text: powershell {"command": "Select-String -Pattern 'TODO' -Path *.go"} (prefer search_replace builtin)`,
			"find":   `Find files: powershell {"command": "Get-ChildItem -Recurse -Name '*.txt'"}`,
			"cwd":    `Current dir: powershell {"command": "Get-Location"}`,
			"env":    `Environment: powershell {"command": "$env:PATH"}`,
		}
	case "darwin", "linux":
		return map[string]string{
			"list":   `List files: bash {"command": "ls -la *.go"}`,
			"view":   `View file: bash {"command": "cat README.md"} (prefer view builtin)`,
			"write":  `Write file: bash {"command": "echo 'hello' > test.txt"} (prefer create builtin)`,
			"run":    `Run command: bash {"command": "go test ./..."}`,
			"search": `Search text: bash {"command": "grep 'TODO' *.go"} (prefer search_replace builtin)`,
			"find":   `Find files: bash {"command": "find . -name '*.txt'"}`,
			"cwd":    `Current dir: bash {"command": "pwd"}`,
			"env":    `Environment: bash {"command": "echo $PATH"}`,
		}
	default:
		return map[string]string{
			"list":   `List files: Use platform-specific listing command`,
			"view":   `View file: Use view builtin (preferred)`,
			"write":  `Write file: Use create builtin (preferred)`,
			"run":    `Run command: Use platform-specific execution command`,
			"search": `Search text: Use search_replace builtin (preferred)`,
			"find":   `Find files: Use platform-specific file finding command`,
			"cwd":    `Current dir: Use platform-specific directory command`,
			"env":    `Environment: Use platform-specific environment command`,
		}
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// InjectPlatformContext adds OS-specific guidance to agent prompts with filtered commands
var platformCache struct {
	sync.Mutex
	value string
	key   string
}

func InjectPlatformContext(prompt string, allowedCommands []string, allowedBuiltins []string) string {
	const start = "<!-- PLATFORM_CONTEXT_START -->"
	const end = "<!-- PLATFORM_CONTEXT_END -->"
	if strings.Contains(prompt, start) {
		return prompt // already has context
	}
	cacheKey := strings.Join(allowedCommands, ",") + "|" + strings.Join(allowedBuiltins, ",")
	platformCache.Lock()
	if platformCache.key != cacheKey || platformCache.value == "" {
		platformCache.value = GetPlatformContext(allowedCommands, allowedBuiltins)
		platformCache.key = cacheKey
	}
	platformInfo := platformCache.value
	platformCache.Unlock()
	return prompt + "\n" + start + "\n" + platformInfo + end
}

// InjectPlatformContextFromRegistry adds OS + tool guidance using the live registry.
// It injects only once (recognized by PLATFORM_CONTEXT_START/END markers).
func InjectPlatformContextFromRegistry(prompt string, reg tool.Registry) string {
	const start = "<!-- PLATFORM_CONTEXT_START -->"
	const end = "<!-- PLATFORM_CONTEXT_END -->"
	if strings.Contains(prompt, start) {
		return prompt
	}
	// OS header
	var platformInfo string
	switch runtime.GOOS {
	case "windows":
		platformInfo = "PLATFORM: Windows with PowerShell"
	case "darwin":
		platformInfo = "PLATFORM: macOS with Unix shell"
	case "linux":
		platformInfo = "PLATFORM: Linux with Unix shell"
	default:
		platformInfo = "PLATFORM: Unknown OS"
	}

	// Categorize a compact set of builtins from the actual registry
	fileSet := map[string]bool{
		"read_lines": true, "edit_range": true, "insert_at": true, "search_replace": true,
		"fileinfo": true, "view": true, "create": true, "patch": true,
	}
	webSet := map[string]bool{
		"web_search": true, "read_webpage": true, "api": true, "download": true, "fetch": true, "mcp": true,
	}

	var fileTools, webTools, otherTools []string
	// Collect tool names from registry
	for name := range reg {
		if fileSet[name] {
			fileTools = append(fileTools, name)
			continue
		}
		if webSet[name] {
			webTools = append(webTools, name)
			continue
		}
		otherTools = append(otherTools, name)
	}
	sort.Strings(fileTools)
	sort.Strings(webTools)
	sort.Strings(otherTools)

	// If tool list injection is disabled, provide only OS header
	if env.Bool("AGENTRY_DISABLE_TOOL_LIST", true) {
		return prompt + "\n" + start + "\n" + platformInfo + end
	}

	// Limit list sizes to keep context minimal
	capList := func(in []string, n int) []string {
		if n <= 0 || len(in) <= n {
			return in
		}
		return in[:n]
	}
	fileTools = capList(fileTools, 8)
	webTools = capList(webTools, 6)
	otherTools = capList(otherTools, 6)

	var b strings.Builder
	b.WriteString(platformInfo)
	b.WriteString("\n\n")
	if len(fileTools)+len(webTools)+len(otherTools) > 0 {
		b.WriteString("🎯 PREFERRED TOOLS (discovery from registry):\n")
		if len(fileTools) > 0 {
			b.WriteString("\n📁 File Operations:\n")
			for _, n := range fileTools {
				if t, ok := reg[n]; ok {
					b.WriteString("- ")
					b.WriteString(n)
					desc := t.Description()
					if desc != "" {
						b.WriteString(": ")
						b.WriteString(desc)
					}
					b.WriteString("\n")
				}
			}
		}
		if len(webTools) > 0 {
			b.WriteString("\n🌐 Web & Network:\n")
			for _, n := range webTools {
				if t, ok := reg[n]; ok {
					b.WriteString("- ")
					b.WriteString(n)
					desc := t.Description()
					if desc != "" {
						b.WriteString(": ")
						b.WriteString(desc)
					}
					b.WriteString("\n")
				}
			}
		}
		if len(otherTools) > 0 {
			b.WriteString("\n🔧 Other Tools:\n")
			for _, n := range otherTools {
				if t, ok := reg[n]; ok {
					b.WriteString("- ")
					b.WriteString(n)
					desc := t.Description()
					if desc != "" {
						b.WriteString(": ")
						b.WriteString(desc)
					}
					b.WriteString("\n")
				}
			}
		}
	}

	return prompt + "\n" + start + "\n" + b.String() + end
}

// InjectAvailableRoles adds available agent role information to the prompt
func InjectAvailableRoles(prompt string, availableRoles []string) string {
	// If prompt already contains role info, don't duplicate
	if strings.Contains(prompt, "AVAILABLE AGENTS:") {
		return prompt
	}

	if len(availableRoles) == 0 {
		return prompt
	}

    // Ensure deterministic order for display
    names := append([]string(nil), availableRoles...)
    sort.Strings(names)

    var roleInfo strings.Builder
    roleInfo.WriteString("\n\nAVAILABLE AGENTS: You can delegate tasks to these specialized agents using the 'agent' tool:\n")
    for _, role := range names {
        if role != "agent_0" { // Don't list ourselves
            roleInfo.WriteString("- ")
            roleInfo.WriteString(role)
            roleInfo.WriteString("\n")
        }
    }
	roleInfo.WriteString("\nExample delegation: {\"agent\": \"coder\", \"input\": \"create a hello world program\"}")

	return prompt + roleInfo.String()
}
