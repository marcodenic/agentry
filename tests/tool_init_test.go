package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/plugin"
)

func TestToolInitScaffold(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	if err := plugin.InitTool("demo"); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	goFile := filepath.Join("demo", "demo.go")
	yamlFile := filepath.Join("demo", "demo.yaml")

	if _, err := os.Stat(goFile); err != nil {
		t.Fatalf("missing go file: %v", err)
	}
	if _, err := os.Stat(yamlFile); err != nil {
		t.Fatalf("missing yaml file: %v", err)
	}

	gb, _ := os.ReadFile(goFile)
	if !bytes.Contains(gb, []byte("func Exec")) {
		t.Fatalf("exec function not found")
	}
	yb, _ := os.ReadFile(yamlFile)
	if !bytes.Contains(yb, []byte("name: demo")) {
		t.Fatalf("name not written")
	}
}
