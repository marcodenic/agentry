package debug

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogToFileWritesEntry(t *testing.T) {
	tempDir := t.TempDir()
	rl, err := NewRollingLogger(tempDir, "test", 1024)
	if err != nil {
		t.Fatalf("NewRollingLogger: %v", err)
	}
	original := fileLogger
	fileLogger = rl
	t.Cleanup(func() {
		rl.Close()
		fileLogger = original
	})

	LogToFile("INFO", "hello %s", "world")

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected log file to be created")
	}
	data, err := os.ReadFile(filepath.Join(tempDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "[INFO] hello world") {
		t.Fatalf("expected log entry to contain formatted message, got %q", string(data))
	}
}

func TestRollingLoggerRollsWhenExceedingMaxSize(t *testing.T) {
	tempDir := t.TempDir()
	rl, err := NewRollingLogger(tempDir, "rotate", 32)
	if err != nil {
		t.Fatalf("NewRollingLogger: %v", err)
	}
	t.Cleanup(func() { rl.Close() })

	if _, err := rl.Write([]byte(strings.Repeat("a", 24))); err != nil {
		t.Fatalf("initial write failed: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if _, err := rl.Write([]byte(strings.Repeat("b", 24))); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected log rotation to create new file, files=%d", len(entries))
	}
}
