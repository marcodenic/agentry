package tests

import (
	"os"
	"testing"
)

func TestIsGoProject(t *testing.T) {
	// This project contains Go source under cmd/ and internal/; recognize that structure.
	candidates := []string{
		"../cmd/agentry/main.go",
		"../internal/version.go",
		"../go.mod",
	}
	found := false
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find Go project markers (one of %v)", candidates)
	}
}
