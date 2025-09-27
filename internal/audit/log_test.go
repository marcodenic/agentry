package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRequiresPath(t *testing.T) {
	if _, err := Open("", 0); err == nil {
		t.Fatalf("expected error for empty path")
	}
}

func TestLogRotateKeepsLatestFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	l, err := Open(path, int64(len("first\n")))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if l != nil {
			l.Close()
		}
	})

	first := []byte("first\n")
	if _, err := l.Write(first); err != nil {
		t.Fatalf("write first: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read active log: %v", err)
	}
	if string(data) != string(first) {
		t.Fatalf("unexpected contents before rotation: %q", data)
	}

	second := []byte("second entry\n")
	if _, err := l.Write(second); err != nil {
		t.Fatalf("write second: %v", err)
	}

	if err := l.Close(); err != nil {
		t.Fatalf("close log: %v", err)
	}
	l = nil

	files, err := filepath.Glob(path + ".*")
	if err != nil {
		t.Fatalf("glob rotated files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one rotated file, got %d (%v)", len(files), files)
	}

	rotated, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("read rotated: %v", err)
	}
	if string(rotated) != string(first) {
		t.Fatalf("rotated contents mismatch: %q", rotated)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if string(current) != string(second) {
		t.Fatalf("current contents mismatch: %q", current)
	}
}
