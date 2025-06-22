package lsp

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

var languages []string

func init() {
	if _, err := os.Stat("go.mod"); err == nil {
		languages = append(languages, "go")
	}
	if _, err := os.Stat("tsconfig.json"); err == nil {
		languages = append(languages, "typescript")
	}
	if _, err := os.Stat(filepath.Join("ts-sdk", "tsconfig.json")); err == nil {
		if !contains("typescript", languages) {
			languages = append(languages, "typescript")
		}
	}
}

func Languages() []string { return languages }

func contains(lang string, langs []string) bool {
	for _, l := range langs {
		if l == lang {
			return true
		}
	}
	return false
}

// Check runs language server diagnostics on the provided files.
func Check(files []string) (string, error) {
	var out bytes.Buffer
	var goFiles, tsFiles []string
	for _, f := range files {
		switch filepath.Ext(f) {
		case ".go":
			goFiles = append(goFiles, f)
		case ".ts", ".tsx":
			tsFiles = append(tsFiles, f)
		}
	}
	if len(goFiles) > 0 && contains("go", languages) {
		args := append([]string{"check"}, goFiles...)
		cmd := exec.Command("gopls", args...)
		b, err := cmd.CombinedOutput()
		out.Write(b)
		if err != nil {
			return out.String(), err
		}
	}
	if len(tsFiles) > 0 && contains("typescript", languages) {
		args := append([]string{"--noEmit"}, tsFiles...)
		cmd := exec.Command("tsc", args...)
		b, err := cmd.CombinedOutput()
		out.Write(b)
		if err != nil {
			return out.String(), err
		}
	}
	return out.String(), nil
}
