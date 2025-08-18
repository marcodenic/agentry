package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestLSPDiagnosticsBuiltin(t *testing.T) {
	// Skip if gopls is not installed; CI may not have it
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls not installed; skipping LSP diagnostics test")
	}

	// Create temp workspace
	dir := t.TempDir()
	goMod := []byte("module example.com/tmp\n\ngo 1.23\n")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goMod, 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	// Intentionally create a tiny valid Go file
	mainGo := []byte("package main\nfunc main(){}\n")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), mainGo, 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	// Switch cwd
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)

	reg := tool.DefaultRegistry()
	ttool, ok := reg.Use("lsp_diagnostics")
	if !ok {
		t.Skip("lsp_diagnostics tool not registered")
	}

	out, err := ttool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("lsp_diagnostics failed: %v; out=%s", err, out)
	}
	if !strings.Contains(out, "languages") {
		t.Errorf("expected languages key in output, got: %s", out)
	}
}
