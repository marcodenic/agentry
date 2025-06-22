package patch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	diff "github.com/sourcegraph/go-diff/diff"
)

// Change represents a single file changed by a patch.
type Change struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// Result contains metadata about an applied patch.
type Result struct {
	Files []Change `json:"files"`
}

// Apply parses a unified diff and applies it to the local filesystem.
// It returns metadata about the applied changes.
func Apply(patchStr string) (Result, error) {
	fds, err := diff.ParseMultiFileDiff([]byte(patchStr))
	if err != nil {
		return Result{}, err
	}
	var res Result
	for _, fd := range fds {
		path := choosePath(fd)
		st := fd.Stat()
		change := Change{Path: path, Additions: int(st.Added), Deletions: int(st.Deleted)}
		if fd.NewName == "/dev/null" {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return res, err
			}
			res.Files = append(res.Files, change)
			continue
		}
		var lines []string
		if fd.OrigName != "/dev/null" {
			b, err := os.ReadFile(path)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return res, err
			}
			lines = splitLines(string(b))
		}
		patched, err := applyHunks(lines, fd.Hunks)
		if err != nil {
			return res, fmt.Errorf("%s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return res, err
		}
		content := strings.Join(patched, "\n")
		if len(patched) > 0 {
			content += "\n"
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return res, err
		}
		res.Files = append(res.Files, change)
	}
	return res, nil
}

func choosePath(fd *diff.FileDiff) string {
	p := fd.NewName
	if p == "/dev/null" {
		p = fd.OrigName
	}
	p = strings.TrimPrefix(p, "a/")
	p = strings.TrimPrefix(p, "b/")
	return p
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

func applyHunks(orig []string, hunks []*diff.Hunk) ([]string, error) {
	out := make([]string, 0, len(orig))
	idx := 0
	for _, h := range hunks {
		start := int(h.OrigStartLine) - 1
		for idx < start && idx < len(orig) {
			out = append(out, orig[idx])
			idx++
		}
		bodyLines := bytes.Split(bytes.TrimSuffix(h.Body, []byte{'\n'}), []byte{'\n'})
		for _, b := range bodyLines {
			if len(b) == 0 {
				continue
			}
			prefix := b[0]
			line := string(b[1:])
			switch prefix {
			case ' ':
				if idx >= len(orig) || orig[idx] != line {
					return nil, fmt.Errorf("context mismatch: expected %q, got %q", origLine(orig, idx), line)
				}
				out = append(out, line)
				idx++
			case '-':
				if idx >= len(orig) || orig[idx] != line {
					return nil, fmt.Errorf("delete mismatch: expected %q, got %q", origLine(orig, idx), line)
				}
				idx++
			case '+':
				out = append(out, line)
			case '\\':
				// ignore no newline marker
			}
		}
	}
	for idx < len(orig) {
		out = append(out, orig[idx])
		idx++
	}
	return out, nil
}

func origLine(orig []string, idx int) string {
	if idx >= 0 && idx < len(orig) {
		return orig[idx]
	}
	return ""
}

// MarshalResult marshals a Result to JSON.
func MarshalResult(r Result) (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
