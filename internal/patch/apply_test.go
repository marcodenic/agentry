package patch

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempDir(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() {
		_ = os.Chdir(cwd)
	}
}

func TestApplyCreatesAndModifiesFiles(t *testing.T) {
	defer withTempDir(t)()

	createPatch := "" +
		"diff --git a/notes.txt b/notes.txt\n" +
		"new file mode 100644\n" +
		"index 0000000..1111111\n" +
		"--- /dev/null\n" +
		"+++ b/notes.txt\n" +
		"@@ -0,0 +1,2 @@\n" +
		"+hello\n" +
		"+world\n"

	res, err := Apply(createPatch)
	if err != nil {
		t.Fatalf("apply create patch: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0].Path != "notes.txt" || res.Files[0].Additions != 2 {
		t.Fatalf("unexpected result for create: %+v", res)
	}

	content, err := os.ReadFile("notes.txt")
	if err != nil {
		t.Fatalf("read notes: %v", err)
	}
	if string(content) != "hello\nworld\n" {
		t.Fatalf("unexpected file contents: %q", content)
	}

	modifyPatch := "" +
		"diff --git a/notes.txt b/notes.txt\n" +
		"index 1111111..2222222 100644\n" +
		"--- a/notes.txt\n" +
		"+++ b/notes.txt\n" +
		"@@ -1,2 +1,2 @@\n" +
		"-hello\n" +
		"-world\n" +
		"+hello\n" +
		"+gophers\n"

	res, err = Apply(modifyPatch)
	if err != nil {
		t.Fatalf("apply modify patch: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0].Deletions != 1 || res.Files[0].Additions != 1 {
		t.Fatalf("unexpected result for modify: %+v", res)
	}

	content, err = os.ReadFile("notes.txt")
	if err != nil {
		t.Fatalf("read notes after modify: %v", err)
	}
	if string(content) != "hello\ngophers\n" {
		t.Fatalf("unexpected content after modify: %q", content)
	}

	deletePatch := "" +
		"diff --git a/notes.txt b/notes.txt\n" +
		"deleted file mode 100644\n" +
		"index 2222222..0000000\n" +
		"--- a/notes.txt\n" +
		"+++ /dev/null\n" +
		"@@ -1,2 +0,0 @@\n" +
		"-hello\n" +
		"-gophers\n"

	res, err = Apply(deletePatch)
	if err != nil {
		t.Fatalf("apply delete patch: %v", err)
	}
	if len(res.Files) != 1 || res.Files[0].Deletions != 2 {
		t.Fatalf("unexpected result for delete: %+v", res)
	}
	if _, err := os.Stat("notes.txt"); !os.IsNotExist(err) {
		t.Fatalf("expected notes.txt removed, err=%v", err)
	}
}

func TestApplyContextMismatch(t *testing.T) {
	defer withTempDir(t)()

	if err := os.WriteFile("sample.txt", []byte("first\n"), 0644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	mismatch := "" +
		"diff --git a/sample.txt b/sample.txt\n" +
		"index 1111111..3333333 100644\n" +
		"--- a/sample.txt\n" +
		"+++ b/sample.txt\n" +
		"@@ -1 +1 @@\n" +
		"-second\n" +
		"+third\n"

	if _, err := Apply(mismatch); err == nil {
		t.Fatalf("expected context mismatch error")
	}
}

func TestMarshalResult(t *testing.T) {
	res := Result{Files: []Change{{Path: filepath.ToSlash("a/b"), Additions: 1}}}
	out, err := MarshalResult(res)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if out == "" || out[0] != '{' {
		t.Fatalf("unexpected json output: %q", out)
	}
}
