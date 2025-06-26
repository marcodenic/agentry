package tests

import (
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestSemanticCommandSystem(t *testing.T) {
	// Create test agent
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: staticClient{out: "test response"}}}
	parent := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	// Create team context
	team, err := converse.NewTeamContext(parent)
	if err != nil {
		t.Fatalf("NewTeamContext failed: %v", err)
	}

	// Add coder agent (should use new semantic command system)
	agent, name := team.AddAgent("coder")
	if agent == nil {
		t.Error("AddAgent should return a valid agent")
	}
	if name != "coder" {
		t.Errorf("Expected agent name 'coder', got '%s'", name)
	}

	// Check that agent prompt contains platform-specific guidance
	if agent.Prompt == "" {
		t.Error("Agent prompt should not be empty")
	}

	// Check that platform context was injected
	if !stringContains(agent.Prompt, "PLATFORM:") {
		t.Error("Agent prompt should contain platform context")
	}
	if !stringContains(agent.Prompt, "ALLOWED COMMANDS:") {
		t.Error("Agent prompt should contain allowed commands section")
	}
	if !stringContains(agent.Prompt, "BUILTIN TOOLS:") {
		t.Error("Agent prompt should contain builtin tools section")
	}

	// On Windows, should show PowerShell commands
	if stringContains(agent.Prompt, "PLATFORM: Windows") {
		if !stringContains(agent.Prompt, "powershell") {
			t.Error("Windows platform should show PowerShell commands")
		}
		if !stringContains(agent.Prompt, "Get-Content") {
			t.Error("Windows platform should show Get-Content for view command")
		}
	}

	t.Logf("Agent prompt (first 500 chars): %s", truncate(agent.Prompt, 500))
}

func stringContains(text, substr string) bool {
	return len(text) >= len(substr) && findInString(text, substr)
}

func findInString(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
