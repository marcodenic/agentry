package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func init() {
	// Register simple write builtin
	builtinMap["write"] = builtinSpec{
		Desc: "Write full content to a file (create or overwrite)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file":   map[string]any{"type": "string", "description": "File path (alias: path)"},
				"path":   map[string]any{"type": "string", "description": "File path (alias: file)"},
				"content": map[string]any{"type": "string", "description": "Content to write (alias: text)"},
				"text":    map[string]any{"type": "string", "description": "Content to write (alias: content)"},
			},
			"required": []string{},
			"example":  map[string]any{"file": "test.txt", "content": "hello"},
		},
		Exec: writeFileExec,
	}

	// Register simple edit builtin
	builtinMap["edit"] = builtinSpec{
		Desc: "Edit (overwrite) a file's content; requires prior view and unchanged file",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file":   map[string]any{"type": "string", "description": "File path (alias: path)"},
				"path":   map[string]any{"type": "string", "description": "File path (alias: file)"},
				"content": map[string]any{"type": "string", "description": "New content (alias: text)"},
				"text":    map[string]any{"type": "string", "description": "New content (alias: content)"},
			},
			"required": []string{},
			"example":  map[string]any{"path": "test.txt", "text": "updated"},
		},
		Exec: editFileExec,
	}
}

func getPathAndContent(args map[string]any) (path string, content string) {
	if p, ok := args["file"].(string); ok && p != "" {
		path = p
	}
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}
	if c, ok := args["content"].(string); ok {
		content = c
	}
	if c, ok := args["text"].(string); ok && c != "" {
		content = c
	}
	return
}

func writeFileExec(ctx context.Context, args map[string]any) (string, error) {
	p, content := getPathAndContent(args)
	if p == "" {
		return "", errors.New("missing path")
	}
	p = absPath(p)

	// If we have a prior view record and file exists, ensure unchanged since view
	if v, ok := viewedFiles.Load(p); ok {
		if info, err := os.Stat(p); err == nil {
			if ts, ok2 := v.(time.Time); ok2 {
				if !info.ModTime().Equal(ts) {
					return "", fmt.Errorf("file changed on disk since view: %s", p)
				}
			}
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	_ = recordView(p)

	out := map[string]any{"path": p, "bytes": len(content)}
	b, _ := json.Marshal(out)
	return string(b), nil
}

func editFileExec(ctx context.Context, args map[string]any) (string, error) {
	p, content := getPathAndContent(args)
	if p == "" {
		return "", errors.New("missing path")
	}
	p = absPath(p)

	// Require prior view
	v, ok := viewedFiles.Load(p)
	if !ok {
		return "", errors.New("cannot edit without prior view")
	}
	if info, err := os.Stat(p); err == nil {
		if ts, ok2 := v.(time.Time); ok2 {
			if !info.ModTime().Equal(ts) {
				return "", fmt.Errorf("file changed on disk since view: %s", p)
			}
		}
	} else {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}
	_ = recordView(p)

	out := map[string]any{"path": p, "edited": true, "bytes": len(content)}
	b, _ := json.Marshal(out)
	return string(b), nil
}
