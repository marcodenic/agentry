package core

import (
	"runtime"
	"strings"
)

// GetPlatformContext returns OS-specific guidance for agents with filtered commands
func GetPlatformContext(allowedCommands []string, allowedBuiltins []string) string {
	// Define command mappings for each OS
	var commandMap map[string]string
	var platformInfo string
	
	switch runtime.GOOS {
	case "windows":
		platformInfo = "PLATFORM: Windows with PowerShell\nSHELL TOOLS: powershell, cmd"
		commandMap = map[string]string{
			"list":   `List files: powershell {"command": "Get-ChildItem -Name '*.go'"}`,
			"view":   `View file: powershell {"command": "Get-Content README.md"}`,
			"write":  `Write file: powershell {"command": "Set-Content -Path test.txt -Value 'hello'"}`,
			"run":    `Run command: powershell {"command": "go test ./..."}`,
			"search": `Search text: powershell {"command": "Select-String -Pattern 'TODO' -Path *.go"}`,
			"find":   `Find files: powershell {"command": "Get-ChildItem -Recurse -Name '*.txt'"}`,
			"cwd":    `Current dir: powershell {"command": "Get-Location"}`,
			"env":    `Environment: powershell {"command": "$env:PATH"}`,
		}
	case "darwin":
		platformInfo = "PLATFORM: macOS with Unix shell\nSHELL TOOLS: bash, sh"
		commandMap = map[string]string{
			"list":   `List files: bash {"command": "ls -la *.go"}`,
			"view":   `View file: bash {"command": "cat README.md"}`,
			"write":  `Write file: bash {"command": "echo 'hello' > test.txt"}`,
			"run":    `Run command: bash {"command": "go test ./..."}`,
			"search": `Search text: bash {"command": "grep 'TODO' *.go"}`,
			"find":   `Find files: bash {"command": "find . -name '*.txt'"}`,
			"cwd":    `Current dir: bash {"command": "pwd"}`,
			"env":    `Environment: bash {"command": "echo $PATH"}`,
		}
	case "linux":
		platformInfo = "PLATFORM: Linux with Unix shell\nSHELL TOOLS: bash, sh"
		commandMap = map[string]string{
			"list":   `List files: bash {"command": "ls -la *.go"}`,
			"view":   `View file: bash {"command": "cat README.md"}`,
			"write":  `Write file: bash {"command": "echo 'hello' > test.txt"}`,
			"run":    `Run command: bash {"command": "go test ./..."}`,
			"search": `Search text: bash {"command": "grep 'TODO' *.go"}`,
			"find":   `Find files: bash {"command": "find . -name '*.txt'"}`,
			"cwd":    `Current dir: bash {"command": "pwd"}`,
			"env":    `Environment: bash {"command": "echo $PATH"}`,
		}
	default:
		platformInfo = "PLATFORM: Unknown OS\nSHELL TOOLS: Use fetch, agent, ping, echo tools when possible"
		commandMap = map[string]string{
			"list":   `List files: Use platform-specific listing command`,
			"view":   `View file: Use platform-specific file reading command`,
			"write":  `Write file: Use platform-specific file writing command`,
			"run":    `Run command: Use platform-specific execution command`,
			"search": `Search text: Use platform-specific text search command`,
			"find":   `Find files: Use platform-specific file finding command`,
			"cwd":    `Current dir: Use platform-specific directory command`,
			"env":    `Environment: Use platform-specific environment command`,
		}
	}
	
	result := platformInfo + "\n"
	
	// Add allowed commands section
	if len(allowedCommands) > 0 {
		result += "\nALLOWED COMMANDS:\n"
		for _, cmd := range allowedCommands {
			if cmdExample, exists := commandMap[cmd]; exists {
				result += "- " + cmdExample + "\n"
			}
		}
	}
	
	// Add builtin tools section
	if len(allowedBuiltins) > 0 {
		result += "\nBUILTIN TOOLS:\n"
		for _, builtin := range allowedBuiltins {
			result += "- " + builtin + "\n"
		}
	}
	
	return result
}

// GetPlatformContextLegacy returns the old format for backward compatibility
func GetPlatformContextLegacy() string {
	return GetPlatformContext(
		[]string{"list", "view", "write", "run", "search", "find", "cwd", "env"},
		[]string{},
	)
}

// InjectPlatformContext adds OS-specific guidance to agent prompts with filtered commands
func InjectPlatformContext(prompt string, allowedCommands []string, allowedBuiltins []string) string {
	platformInfo := GetPlatformContext(allowedCommands, allowedBuiltins)
	
	// If prompt already contains platform info, don't duplicate
	if strings.Contains(prompt, "PLATFORM:") {
		return prompt
	}
	
	return prompt + "\n" + platformInfo
}

// InjectPlatformContextLegacy provides backward compatibility
func InjectPlatformContextLegacy(prompt string) string {
	return InjectPlatformContext(prompt, 
		[]string{"list", "view", "write", "run", "search", "find", "cwd", "env"},
		[]string{},
	)
}
