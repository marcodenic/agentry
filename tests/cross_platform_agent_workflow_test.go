package tests

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// Simple client for testing cross-platform agent workflow without delegation
type simpleCrossPlatformClient struct{}

func (c simpleCrossPlatformClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	if len(msgs) == 0 {
		return model.Completion{Content: "No messages provided"}, nil
	}
	
	lastMsg := msgs[len(msgs)-1].Content
	
	// If requesting file creation and directory listing
	if strings.Contains(lastMsg, "create a file named 'platform_test.txt'") {
		return model.Completion{
			Content: "I'll create the file and then list the directory contents as requested.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_1",
					Name: "write",
					Arguments: []byte(`{"file": "platform_test.txt", "content": "Platform: ` + runtime.GOOS + `"}`),
				},
			},
		}, nil
	}
	
	// If file creation is confirmed, list directory
	if strings.Contains(lastMsg, "powershell.exe") || strings.Contains(lastMsg, "echo") || strings.Contains(lastMsg, "Set-Content") || strings.Contains(lastMsg, "File created") {
		return model.Completion{
			Content: "File created successfully. Now I'll list the directory contents.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_2",
					Name: "ls",
					Arguments: []byte(`{"path": "."}`),
				},
			},
		}, nil
	}
	
	return model.Completion{Content: "Task completed successfully."}, nil
}

func TestCrossPlatformAgentWorkflow(t *testing.T) {
	// Test that an agent can use cross-platform tools in a realistic scenario
	ctx := context.Background()
	
	// Set sandbox to disabled for direct execution
	tool.SetSandboxEngine("disabled")

	// Create a temporary working directory
	tempDir, err := os.MkdirTemp("", "agentry-agent-crossplatform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Get tool registry
	registry := tool.DefaultRegistry()

	// Create a single agent with a simple mock client that just executes tools
	route := router.Rules{{Name: "crossplatform", IfContains: []string{""}, Client: simpleCrossPlatformClient{}}}
	agent := core.New(route, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	agent.Prompt = `You are a file management agent. You have access to file system tools like write, view, ls, edit, etc.`

	// Create test team
	tm, err := converse.NewTeam(agent, 1, "test")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}
	
	ctx = team.WithContext(ctx, tm)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Test agent workflow: create a file and list directory
	prompt := fmt.Sprintf("Please create a file named 'platform_test.txt' with content 'Platform: %s' and then list the directory contents", runtime.GOOS)
	
	response, err := agent.Run(ctx, prompt)
	if err != nil {
		t.Fatalf("Agent workflow failed: %v", err)
	}

	t.Logf("Agent workflow response: %s", response)

	// Verify that the file was created
	if _, err := os.Stat("platform_test.txt"); os.IsNotExist(err) {
		t.Error("Agent workflow did not create the expected file")
	}

	// Verify content is correct
	content, err := os.ReadFile("platform_test.txt")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expectedContent := fmt.Sprintf("Platform: %s", runtime.GOOS)
	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("File content incorrect. Expected to contain '%s', got: %s", expectedContent, string(content))
	}

	t.Logf("Successfully created file with correct platform-specific content: %s", string(content))
}
