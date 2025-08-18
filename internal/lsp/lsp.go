package lsp

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	// Python projects: pyproject.toml or requirements.txt
	if _, err := os.Stat("pyproject.toml"); err == nil {
		languages = append(languages, "python")
	} else if _, err := os.Stat("requirements.txt"); err == nil {
		languages = append(languages, "python")
	}
	// Rust projects: Cargo.toml
	if _, err := os.Stat("Cargo.toml"); err == nil {
		languages = append(languages, "rust")
	}
	// JavaScript projects: package.json (without tsconfig.json)
	if _, err := os.Stat("package.json"); err == nil {
		if !contains("typescript", languages) {
			languages = append(languages, "javascript")
		} else {
			// Many TS projects also have JS files; include JS as well
			languages = append(languages, "javascript")
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

// Diagnostic represents a single compiler/language diagnostic.
type Diagnostic struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Code     string `json:"code,omitempty"`
	Severity string `json:"severity"` // error|warning|info
	Message  string `json:"message"`
	Tool     string `json:"tool"`     // gopls|tsc
	Language string `json:"language"` // go|typescript
}

// Check runs language server diagnostics on the provided files.
func Check(files []string) (string, error) {
	var out bytes.Buffer
	var goFiles, tsFiles, pyFiles, rsFiles, jsFiles []string
	for _, f := range files {
		switch filepath.Ext(f) {
		case ".go":
			goFiles = append(goFiles, f)
		case ".ts", ".tsx":
			tsFiles = append(tsFiles, f)
		case ".py":
			pyFiles = append(pyFiles, f)
		case ".rs":
			rsFiles = append(rsFiles, f)
		case ".js", ".jsx", ".mjs", ".cjs":
			jsFiles = append(jsFiles, f)
		}
	}
	if len(goFiles) > 0 && contains("go", languages) {
		if _, err := exec.LookPath("gopls"); err == nil {
			args := append([]string{"check"}, goFiles...)
			cmd := exec.Command("gopls", args...)
			b, err := cmd.CombinedOutput()
			out.Write(b)
			if err != nil {
				return out.String(), err
			}
		}
	}
	if len(tsFiles) > 0 && contains("typescript", languages) {
		if _, err := exec.LookPath("tsc"); err == nil {
			args := append([]string{"--noEmit"}, tsFiles...)
			cmd := exec.Command("tsc", args...)
			b, err := cmd.CombinedOutput()
			out.Write(b)
			if err != nil {
				return out.String(), err
			}
		}
	}
	if len(pyFiles) > 0 && contains("python", languages) {
		if _, err := exec.LookPath("pyright"); err == nil {
			args := append([]string{}, pyFiles...)
			cmd := exec.Command("pyright", args...)
			b, err := cmd.CombinedOutput()
			out.Write(b)
			if err != nil {
				return out.String(), err
			}
		}
	}
	if len(rsFiles) > 0 && contains("rust", languages) {
		if _, err := exec.LookPath("cargo"); err == nil {
			// cargo check will analyze the entire crate; ignore file list
			cmd := exec.Command("cargo", "check")
			b, err := cmd.CombinedOutput()
			out.Write(b)
			if err != nil {
				return out.String(), err
			}
		}
	}
	if len(jsFiles) > 0 && contains("javascript", languages) {
		if _, err := exec.LookPath("eslint"); err == nil {
			args := append([]string{"--no-error-on-unmatched-pattern"}, jsFiles...)
			cmd := exec.Command("eslint", args...)
			b, err := cmd.CombinedOutput()
			out.Write(b)
			if err != nil {
				return out.String(), err
			}
		}
	}
	return out.String(), nil
}

// ParseDiagnostics attempts to parse output from gopls check and tsc --noEmit
// into structured diagnostics. It is resilient to slight format variations.
func ParseDiagnostics(output string) []Diagnostic {
	var diags []Diagnostic
	if output == "" {
		return diags
	}
	lines := bytes.Split([]byte(output), []byte("\n"))

	// tsc formats:
	// 1) path.ts(12,5): error TS1234: Message
	// 2) path.ts:12:5 - error TS1234: Message
	reTSC1 := regexp.MustCompile(`^(?P<file>.+?)\((?P<line>\d+),(?P<col>\d+)\):\s*(?P<sev>error|warning)\s+(?P<code>TS\d+):\s*(?P<msg>.+)$`)
	reTSC2 := regexp.MustCompile(`^(?P<file>.+?):(?P<line>\d+):(?P<col>\d+)\s*-\s*(?P<sev>error|warning)\s+(?P<code>TS\d+):\s*(?P<msg>.+)$`)

	// gopls format (typical): path.go:12:5: message (may include prefix like warning: )
	reGo := regexp.MustCompile(`^(?P<file>.+?):(?P<line>\d+):(?P<col>\d+):\s*(?P<prefix>(?:warning|error)?:\s*)?(?P<msg>.+)$`)

	// pyright format: path.py:12:5 - error|warning|information: Message [code]
	rePyright := regexp.MustCompile(`^(?P<file>.+?\.py):(?P<line>\d+):(?P<col>\d+)\s*-\s*(?P<sev>error|warning|information):\s*(?P<msg>.+?)(?:\s*\[(?P<code>[^\]]+)\])?$`)

	// eslint stylish: header 'path.js' then lines '  3:10  error  Message  rule'
	reESLintHeader := regexp.MustCompile(`^\s*(?P<file>.+?\.(?:js|jsx|mjs|cjs|ts|tsx))\s*$`)
	reESLintLine := regexp.MustCompile(`^\s*(?P<line>\d+):(?P<col>\d+)\s+(?P<sev>error|warning)\s+(?P<msg>.+?)\s+(?P<code>[a-z0-9\-]+)\s*$`)

	// cargo check fragment: ' --> path.rs:12:5' captured with previous severity
	reCargoLoc := regexp.MustCompile(`^\s*-->\s+(?P<file>.+?\.rs):(?P<line>\d+):(?P<col>\d+)\s*$`)

	var eslintFile string
	lastSeverity := "error" // used for rust cargo parsing
	cargoErrRe := regexp.MustCompile(`^error(\[[^\]]+\])?:`)
	cargoWarnRe := regexp.MustCompile(`^warning:`)
	for _, lb := range lines {
		line := string(bytes.TrimSpace(lb))
		if line == "" {
			continue
		}

		// Track cargo severity lines like 'error[E...]:' or 'warning: ...'
		if cargoErrRe.MatchString(line) {
			lastSeverity = "error"
		} else if cargoWarnRe.MatchString(line) {
			lastSeverity = "warning"
		}

		if m := rePyright.FindStringSubmatch(line); m != nil {
			sev := m[rePyright.SubexpIndex("sev")]
			diags = append(diags, Diagnostic{
				File:     m[rePyright.SubexpIndex("file")],
				Line:     atoiSafe(m[rePyright.SubexpIndex("line")]),
				Col:      atoiSafe(m[rePyright.SubexpIndex("col")]),
				Code:     m[rePyright.SubexpIndex("code")],
				Severity: sev,
				Message:  m[rePyright.SubexpIndex("msg")],
				Tool:     "pyright",
				Language: "python",
			})
			continue
		}
		if m := reESLintHeader.FindStringSubmatch(line); m != nil {
			eslintFile = m[reESLintHeader.SubexpIndex("file")]
			continue
		}
		if eslintFile != "" {
			if m := reESLintLine.FindStringSubmatch(line); m != nil {
				diags = append(diags, Diagnostic{
					File:     eslintFile,
					Line:     atoiSafe(m[reESLintLine.SubexpIndex("line")]),
					Col:      atoiSafe(m[reESLintLine.SubexpIndex("col")]),
					Code:     m[reESLintLine.SubexpIndex("code")],
					Severity: m[reESLintLine.SubexpIndex("sev")],
					Message:  m[reESLintLine.SubexpIndex("msg")],
					Tool:     "eslint",
					Language: "javascript",
				})
				continue
			}
		}
		if m := reCargoLoc.FindStringSubmatch(line); m != nil {
			diags = append(diags, Diagnostic{
				File:     m[reCargoLoc.SubexpIndex("file")],
				Line:     atoiSafe(m[reCargoLoc.SubexpIndex("line")]),
				Col:      atoiSafe(m[reCargoLoc.SubexpIndex("col")]),
				Code:     "",
				Severity: lastSeverity,
				Message:  "cargo check reported an issue",
				Tool:     "cargo",
				Language: "rust",
			})
			continue
		}
		if m := reTSC1.FindStringSubmatch(line); m != nil {
			diags = append(diags, Diagnostic{
				File:     m[reTSC1.SubexpIndex("file")],
				Line:     atoiSafe(m[reTSC1.SubexpIndex("line")]),
				Col:      atoiSafe(m[reTSC1.SubexpIndex("col")]),
				Code:     m[reTSC1.SubexpIndex("code")],
				Severity: m[reTSC1.SubexpIndex("sev")],
				Message:  m[reTSC1.SubexpIndex("msg")],
				Tool:     "tsc",
				Language: "typescript",
			})
			continue
		}
		if m := reTSC2.FindStringSubmatch(line); m != nil {
			diags = append(diags, Diagnostic{
				File:     m[reTSC2.SubexpIndex("file")],
				Line:     atoiSafe(m[reTSC2.SubexpIndex("line")]),
				Col:      atoiSafe(m[reTSC2.SubexpIndex("col")]),
				Code:     m[reTSC2.SubexpIndex("code")],
				Severity: m[reTSC2.SubexpIndex("sev")],
				Message:  m[reTSC2.SubexpIndex("msg")],
				Tool:     "tsc",
				Language: "typescript",
			})
			continue
		}
		if m := reGo.FindStringSubmatch(line); m != nil {
			sev := "error"
			if p := m[reGo.SubexpIndex("prefix")]; p != "" {
				// normalize "warning: " or "error: " prefixes
				if regexp.MustCompile(`(?i)warning`).MatchString(p) {
					sev = "warning"
				}
			}
			diags = append(diags, Diagnostic{
				File:     m[reGo.SubexpIndex("file")],
				Line:     atoiSafe(m[reGo.SubexpIndex("line")]),
				Col:      atoiSafe(m[reGo.SubexpIndex("col")]),
				Code:     "",
				Severity: sev,
				Message:  m[reGo.SubexpIndex("msg")],
				Tool:     "gopls",
				Language: "go",
			})
			continue
		}
		// Unparsed lines are ignored
	}
	return diags
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
