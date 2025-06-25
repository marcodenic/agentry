package core

import (
	"runtime"
	"strings"
)

// GetPlatformContext returns OS-specific guidance for agents
func GetPlatformContext() string {
	switch runtime.GOOS {
	case "windows":
		return `
PLATFORM: Windows with PowerShell
SHELL TOOLS: powershell, cmd
EXAMPLES:
- List files: powershell {"command": "Get-ChildItem -Name '*.go'"}
- View file: powershell {"command": "Get-Content README.md"}
- Write file: powershell {"command": "Set-Content -Path test.txt -Value 'hello'"}
- Run tests: powershell {"command": "go test ./..."}
- Search text: powershell {"command": "Select-String -Pattern 'TODO' -Path *.go"}
- Find files: powershell {"command": "Get-ChildItem -Recurse -Name '*.txt'"}
- Current dir: powershell {"command": "Get-Location"}
- Environment: powershell {"command": "$env:PATH"}
`
	case "darwin":
		return `
PLATFORM: macOS with Unix shell
SHELL TOOLS: bash, sh
EXAMPLES:
- List files: bash {"command": "ls -la *.go"}
- View file: bash {"command": "cat README.md"}
- Write file: bash {"command": "echo 'hello' > test.txt"}
- Run tests: bash {"command": "go test ./..."}
- Search text: bash {"command": "grep 'TODO' *.go"}
- Find files: bash {"command": "find . -name '*.txt'"}
- Current dir: bash {"command": "pwd"}
- Environment: bash {"command": "echo $PATH"}
`
	case "linux":
		return `
PLATFORM: Linux with Unix shell
SHELL TOOLS: bash, sh
EXAMPLES:
- List files: bash {"command": "ls -la *.go"}
- View file: bash {"command": "cat README.md"}
- Write file: bash {"command": "echo 'hello' > test.txt"}
- Run tests: bash {"command": "go test ./..."}
- Search text: bash {"command": "grep 'TODO' *.go"}
- Find files: bash {"command": "find . -name '*.txt'"}
- Current dir: bash {"command": "pwd"}
- Environment: bash {"command": "echo $PATH"}
`
	default:
		return `
PLATFORM: Unknown OS
SHELL TOOLS: Use fetch, agent, ping, echo tools when possible
FALLBACK: Try generic shell commands if platform-specific tools are unavailable
`
	}
}

// InjectPlatformContext adds OS-specific guidance to agent prompts
func InjectPlatformContext(prompt string) string {
	platformInfo := GetPlatformContext()
	
	// If prompt already contains platform info, don't duplicate
	if strings.Contains(prompt, "PLATFORM:") {
		return prompt
	}
	
	return prompt + "\n" + platformInfo
}
