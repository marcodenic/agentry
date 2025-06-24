package tests

import (
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/trace"
)

func TestAnalyzeFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "log*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	log := `{"type":"step_start","data":{"Content":"hi"}}
{"type":"final","data":"bye"}`
	if _, err := tmpFile.WriteString(log); err != nil {
		t.Fatal(err)
	}
	if err := tmpFile.Sync(); err != nil {
		t.Fatal(err)
	}

	sum, err := trace.AnalyzeFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if sum.Tokens != 3 {
		t.Fatalf("expected 3 tokens got %d", sum.Tokens)
	}
}
