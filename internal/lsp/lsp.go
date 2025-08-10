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

	for _, lb := range lines {
		line := string(bytes.TrimSpace(lb))
		if line == "" {
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
