package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/patch"
)

func TestApplyPatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.txt")
	if err := os.WriteFile(path, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	patchStr := "--- a/foo.txt\n+++ b/foo.txt\n@@ -1 +1 @@\n-hello\n+hello world\n"
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)
	stats, err := patch.Apply([]byte(patchStr))
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].Additions != 1 || stats[0].Deletions != 1 {
		t.Fatalf("unexpected stats: %#v", stats[0])
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello world\n" {
		t.Fatalf("patch not applied: %q", string(b))
	}
}
