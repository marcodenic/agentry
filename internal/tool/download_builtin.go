package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func downloadSpec() builtinSpec {
	return builtinSpec{
		Desc: "Download files from URLs to local filesystem",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "URL to download",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Local file path to save to",
				},
				"overwrite": map[string]any{
					"type":        "boolean",
					"description": "Whether to overwrite existing files (default: false)",
					"default":     false,
				},
				"max_size": map[string]any{
					"type":        "integer",
					"description": "Maximum file size in MB (default: 100)",
					"minimum":     1,
					"maximum":     1000,
					"default":     100,
				},
			},
			"required": []string{"url", "path"},
			"example": map[string]any{
				"url":       "https://example.com/data.json",
				"path":      "downloads/data.json",
				"overwrite": false,
				"max_size":  50,
			},
		},
		Exec: downloadFileExec,
	}
}

func registerDownloadBuiltins(reg *builtinRegistry) {
	reg.add("download", downloadSpec())
}

func downloadFileExec(ctx context.Context, args map[string]any) (string, error) {
	urlStr, _ := args["url"].(string)
	if urlStr == "" {
		return "", errors.New("missing url")
	}

	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	overwrite, _ := args["overwrite"].(bool)
	maxSizeMB := 100.0
	if ms, ok := args["max_size"].(float64); ok {
		maxSizeMB = ms
	} else if mi, ok := args["max_size"].(int); ok {
		maxSizeMB = float64(mi)
	}

	path = absPath(path)

	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("file %s already exists (use overwrite=true to replace)", path)
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Agentry/1.0 (File Downloader)")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
			if float64(size)/(1024*1024) > maxSizeMB {
				return "", fmt.Errorf("file size exceeds limit %.1fMB", maxSizeMB)
			}
		}
	}

	tempPath := path + ".download"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	maxBytes := int64(maxSizeMB * 1024 * 1024)
	written, err := io.CopyN(tempFile, resp.Body, maxBytes)
	tempFile.Close()
	if err != nil && err != io.EOF {
		os.Remove(tempPath)
		return "", fmt.Errorf("download failed: %w", err)
	}

	if written == maxBytes {
		os.Remove(tempPath)
		return "", fmt.Errorf("file exceeds size limit %.1fMB", maxSizeMB)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	result := map[string]any{
		"url":         urlStr,
		"path":        path,
		"size_bytes":  written,
		"size_mb":     float64(written) / (1024 * 1024),
		"duration_ms": time.Since(start).Milliseconds(),
		"success":     true,
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
