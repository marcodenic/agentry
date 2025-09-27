package lsp

import (
	"strings"
	"testing"
)

func TestParseDiagnostics(t *testing.T) {
	output := strings.Join([]string{
		"main.go:12:5: warning: use of deprecated api",
		"file.ts(3,7): error TS1234: Something bad",
		"utils.ts:8:2 - warning TS0001: Something else",
		"script.py:9:1 - error: wrong indentation [E999]",
		"src/app.jsx",
		"  3:10  error  Unexpected console statement  no-console",
		"error[E0425]: cannot find value `foo` in this scope",
		" --> src/lib.rs:12:5",
	}, "\n")

	diags := ParseDiagnostics(output)
	if len(diags) != 6 {
		for _, d := range diags {
			t.Logf("diag: %+v", d)
		}
		t.Fatalf("expected 6 diagnostics, got %d", len(diags))
	}

	expect := []Diagnostic{
		{File: "main.go", Line: 12, Col: 5, Severity: "warning", Message: "use of deprecated api", Tool: "gopls", Language: "go"},
		{File: "file.ts", Line: 3, Col: 7, Code: "TS1234", Severity: "error", Message: "Something bad", Tool: "tsc", Language: "typescript"},
		{File: "utils.ts", Line: 8, Col: 2, Code: "TS0001", Severity: "warning", Message: "Something else", Tool: "tsc", Language: "typescript"},
		{File: "script.py", Line: 9, Col: 1, Code: "E999", Severity: "error", Message: "wrong indentation", Tool: "pyright", Language: "python"},
		{File: "src/app.jsx", Line: 3, Col: 10, Code: "no-console", Severity: "error", Message: "Unexpected console statement", Tool: "eslint", Language: "javascript"},
		{File: "src/lib.rs", Line: 12, Col: 5, Severity: "error", Message: "cargo check reported an issue", Tool: "cargo", Language: "rust"},
	}

	for i, want := range expect {
		got := diags[i]
		if got.File != want.File || got.Line != want.Line || got.Col != want.Col || got.Severity != want.Severity || got.Message != want.Message || got.Tool != want.Tool || got.Language != want.Language || got.Code != want.Code {
			t.Fatalf("diag %d mismatch: got %+v want %+v", i, got, want)
		}
	}
}

func TestCheckSkipsAbsentLanguages(t *testing.T) {
	orig := languages
	languages = nil
	t.Cleanup(func() { languages = orig })

	out, err := Check([]string{"main.go", "file.ts"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty output, got %q", out)
	}
}
