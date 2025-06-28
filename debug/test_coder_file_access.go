package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// Mock client that simulates what a real LLM would do when asked to read TODO.md
type testCoderClient struct {
	call int
}

func (c *testCoderClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	c.call++
	
	fmt.Printf("🤖 Call %d: Model received request\n", c.call)
	
	// Print the last message to see what the coder is being asked to do
	if len(msgs) > 0 {
		lastMsg := msgs[len(msgs)-1].Content
		fmt.Printf("📝 Last message: %s\n", strings.TrimSpace(lastMsg))
	}
	
	// Print available tools
	fmt.Printf("🔧 Available tools: ")
	for i, tool := range tools {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%s", tool.Name)
	}
	fmt.Printf("\n")
	
	// Simulate what the coder should do - try to read TODO.md
	if c.call == 1 {
		fmt.Printf("🧠 Coder thinking: I need to read TODO.md file\n")
		
		// Try to use the view tool
		return model.Completion{
			Content: "I'll read the TODO.md file to see what tasks are available.",
			ToolCalls: []model.ToolCall{
				{
					ID:   "call_1",
					Name: "view",
					Arguments: []byte(`{"path": "../TODO.md"}`),
				},
			},
		}, nil
	}
	
	// Second call - respond to the tool result
	return model.Completion{
		Content: "I've reviewed the TODO.md file and can see the available tasks.",
	}, nil
}

func main() {
	fmt.Println("🧪 Testing coder agent file access...")
	
	// Check if TODO.md exists in current directory
	fmt.Printf("📁 Current working directory: %s\n", getCurrentDir())
	
	todoPath := "../TODO.md"  // TODO.md is in parent directory
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		fmt.Printf("❌ TODO.md not found at %s\n", todoPath)
		fmt.Printf("📋 Files in current directory:\n")
		listCurrentDir()
	} else {
		fmt.Printf("✅ TODO.md found at %s\n", todoPath)
	}
	
	// Set up the test environment
	registry := tool.DefaultRegistry()
	client := &testCoderClient{}
	route := router.Rules{{Name: "test", IfContains: []string{""}, Client: client}}
	
	// Create Agent 0 (orchestrator)
	agent0 := core.New(route, registry, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	
	// Create team
	tm, err := converse.NewTeam(agent0, 1, "test")
	if err != nil {
		panic(err)
	}
	
	// Add coder agent
	coderAgent, name := tm.AddAgent("coder")
	fmt.Printf("✅ Created coder agent: %s\n", name)
	fmt.Printf("🔧 Coder has %d tools\n", len(coderAgent.Tools))
	
	// Check if coder has view tool
	if viewTool, hasView := coderAgent.Tools["view"]; hasView {
		fmt.Printf("✅ Coder has view tool: %s\n", viewTool.Description())
		
		// Test the view tool directly
		fmt.Printf("\n🔧 Testing view tool directly...\n")
		result, err := viewTool.Execute(context.Background(), map[string]any{"path": "../TODO.md"})
		if err != nil {
			fmt.Printf("❌ Direct view tool test failed: %v\n", err)
		} else {
			fmt.Printf("✅ Direct view tool works! Result length: %d\n", len(result))
			if len(result) > 100 {
				fmt.Printf("📄 First 100 chars: %s...\n", result[:100])
			} else {
				fmt.Printf("📄 Full result: %s\n", result)
			}
		}
	} else {
		fmt.Printf("❌ Coder does NOT have view tool\n")
	}
	
	// Now test through the team delegation
	fmt.Printf("\n🎯 Testing team delegation to coder...\n")
	ctx := team.WithContext(context.Background(), tm)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// This simulates what happens when Agent 0 delegates to coder
	response, err := tm.Call(ctx, "coder", "Please read ../TODO.md and tell me what tasks are available")
	
	if err != nil {
		fmt.Printf("❌ Team delegation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Team delegation succeeded\n")
		fmt.Printf("📝 Response: %s\n", response)
	}
	
	fmt.Printf("\n🔍 Summary:\n")
	fmt.Printf("- Coder agent created: ✅\n")
	fmt.Printf("- Has view tool: %v\n", coderAgent.Tools["view"] != nil)
	fmt.Printf("- View tool works directly: check above\n")
	fmt.Printf("- Team delegation works: check above\n")
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func listCurrentDir() {
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}
	
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("  📁 %s/\n", file.Name())
		} else {
			fmt.Printf("  📄 %s\n", file.Name())
		}
	}
}
