package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/patch"
)

func TestApplyPatchModify(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\nworld\n"), 0644); err != nil {
		t.Fatal(err)
	}
	patchStr := "--- a/a.txt\n+++ b/a.txt\n@@ -1,2 +1,2 @@\n-hello\n+hi\n world\n"
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)
	res, err := patch.Apply(patchStr)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	out, _ := os.ReadFile("a.txt")
	if string(out) != "hi\nworld\n" {
		t.Fatalf("unexpected result: %q", out)
	}
	if len(res.Files) != 1 || res.Files[0].Path != "a.txt" {
		t.Fatalf("unexpected metadata: %#v", res)
	}
}

func TestApplyPatchNewDelete(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "old.txt"), []byte("bye\n"), 0644)
	patchStr := "--- /dev/null\n+++ b/new.txt\n@@ -0,0 +1 @@\n+new\n" +
		"--- a/old.txt\n+++ /dev/null\n@@ -1 +0,0 @@\n-bye\n"
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)
	_, err := patch.Apply(patchStr)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if b, err := os.ReadFile("new.txt"); err != nil || string(b) != "new\n" {
		t.Fatalf("new file: %v %q", err, b)
	}
	if _, err := os.Stat("old.txt"); !os.IsNotExist(err) {
		t.Fatalf("old file not removed")
	}
}
