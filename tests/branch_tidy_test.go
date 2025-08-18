package tests

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestBranchTidyTool(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temporary git repository for testing
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Initialize git repository
	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatal("failed to init git repo:", err)
	}

	// Configure git user (required for commits)
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial commit
	if err := os.WriteFile("test.txt", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "add", "test.txt").Run()
	if err := exec.Command("git", "commit", "-m", "initial").Run(); err != nil {
		t.Fatal("failed to create initial commit:", err)
	}

	// Create some test branches
	testBranches := []string{"feature/test1", "feature/test2", "bugfix/test3"}
	for _, branch := range testBranches {
		if err := exec.Command("git", "checkout", "-b", branch).Run(); err != nil {
			t.Fatalf("failed to create branch %s: %v", branch, err)
		}
		// Switch back to main
		exec.Command("git", "checkout", "main").Run()
	}

	// Get the branch-tidy tool
	registry := tool.DefaultRegistry()
	branchTidy, exists := registry.Use("branch-tidy")
	if !exists {
		t.Skip("branch-tidy tool not available")
	}

	// Test dry-run first
	t.Run("dry-run", func(t *testing.T) {
		result, err := branchTidy.Execute(context.Background(), map[string]any{
			"dry-run": true,
			"force":   false,
		})
		if err != nil {
			t.Fatal("dry-run failed:", err)
		}

		if !strings.Contains(result, "DRY RUN") {
			t.Error("expected dry-run output")
		}

		// Verify branches still exist
		for _, branch := range testBranches {
			if err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run(); err != nil {
				t.Errorf("branch %s should still exist after dry-run", branch)
			}
		}
	})

	// Test actual deletion
	t.Run("delete-branches", func(t *testing.T) {
		result, err := branchTidy.Execute(context.Background(), map[string]any{
			"dry-run": false,
			"force":   true, // Use force to avoid issues with unmerged branches
		})
		if err != nil {
			t.Fatal("branch deletion failed:", err)
		}

		if !strings.Contains(result, "Successfully deleted") {
			t.Error("expected success message")
		}

		// Verify branches are gone
		for _, branch := range testBranches {
			if err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run(); err == nil {
				t.Errorf("branch %s should be deleted", branch)
			}
		}

		// Verify main branch still exists
		if err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/main").Run(); err != nil {
			t.Error("main branch should still exist")
		}
	})
}

func TestBranchTidySchema(t *testing.T) {
	registry := tool.DefaultRegistry()
	branchTidy, exists := registry.Use("branch-tidy")
	if !exists {
		t.Skip("branch-tidy tool not available")
	}

	schema := branchTidy.JSONSchema()
	if schema["type"] != "object" {
		t.Error("expected object type in schema")
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties in schema")
	}

	if _, exists := properties["force"]; !exists {
		t.Error("expected force property in schema")
	}

	if _, exists := properties["dry-run"]; !exists {
		t.Error("expected dry-run property in schema")
	}
}
