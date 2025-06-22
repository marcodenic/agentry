package patch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	diffpkg "github.com/sourcegraph/go-diff/diff"
)

// FileStat reports additions and deletions for a patched file.
type FileStat struct {
	File      string
	Additions int
	Deletions int
}

// Apply parses and applies a unified diff to the working directory.
// It returns stats about each patched file.
func Apply(data []byte) ([]FileStat, error) {
	fds, err := diffpkg.ParseMultiFileDiff(data)
	if err != nil {
		return nil, err
	}
	stats := make([]FileStat, 0, len(fds))
	for _, fd := range fds {
		path := fd.NewName
		if path == "/dev/null" || path == "" {
			path = fd.OrigName
		}
		path = strings.TrimPrefix(path, "a/")
		path = strings.TrimPrefix(path, "b/")
		path = filepath.FromSlash(path)

		origContent := []byte{}
		if fd.OrigName != "/dev/null" && fd.OrigName != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					origContent = []byte{}
				} else {
					return nil, err
				}
			} else {
				origContent = b
			}
		}
		endNL := len(origContent) == 0 || origContent[len(origContent)-1] == '\n'
		lines := []string{}
		if len(origContent) > 0 {
			trimmed := strings.TrimSuffix(string(origContent), "\n")
			lines = strings.Split(trimmed, "\n")
		}

		newLines, err := applyHunks(lines, fd.Hunks)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		out := strings.Join(newLines, "\n")
		if endNL {
			out += "\n"
		}
		if err := os.WriteFile(path, []byte(out), 0644); err != nil {
			return nil, err
		}
		st := fd.Stat()
		stats = append(stats, FileStat{
			File:      path,
			Additions: int(st.Added + st.Changed),
			Deletions: int(st.Deleted + st.Changed),
		})
	}
	return stats, nil
}

func applyHunks(lines []string, hunks []*diffpkg.Hunk) ([]string, error) {
	out := append([]string{}, lines...)
	offset := 0
	for _, h := range hunks {
		idx := int(h.OrigStartLine-1) + offset
		body := strings.Split(strings.TrimSuffix(string(h.Body), "\n"), "\n")
		pos := idx
		for _, l := range body {
			if l == "" {
				continue
			}
			ch := l[0]
			text := l[1:]
			switch ch {
			case ' ':
				if pos >= len(out) || out[pos] != text {
					return nil, fmt.Errorf("context mismatch")
				}
				pos++
			case '-':
				if pos >= len(out) || out[pos] != text {
					return nil, fmt.Errorf("delete mismatch")
				}
				out = append(out[:pos], out[pos+1:]...)
				offset--
			case '+':
				out = append(out[:pos], append([]string{text}, out[pos:]...)...)
				pos++
				offset++
			}
		}
	}
	return out, nil
}
