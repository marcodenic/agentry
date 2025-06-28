package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func init() {
	builtinMap["api"] = builtinSpec{
		Desc: "Make HTTP/REST API requests",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "API endpoint URL",
				},
				"method": map[string]any{
					"type":        "string",
					"description": "HTTP method (GET, POST, PUT, DELETE, PATCH). Default: GET",
					"enum":        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
					"default":     "GET",
				},
				"headers": map[string]any{
					"type":                 "object",
					"description":          "HTTP headers to send",
					"additionalProperties": map[string]any{"type": "string"},
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Request body (JSON string or raw data)",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Request timeout in seconds (default: 30)",
					"minimum":     1,
					"maximum":     300,
					"default":     30,
				},
			},
			"required": []string{"url"},
			"example": map[string]any{
				"url":    "https://api.github.com/repos/golang/go",
				"method": "GET",
				"headers": map[string]any{
					"Accept": "application/json",
				},
			},
		},
		Exec: apiRequestExec,
	}
}

func apiRequestExec(ctx context.Context, args map[string]any) (string, error) {
	urlStr, _ := args["url"].(string)
	if urlStr == "" {
		return "", errors.New("missing url")
	}

	method := "GET"
	if m, ok := args["method"].(string); ok {
		method = strings.ToUpper(m)
	}

	timeout := 30
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	var body io.Reader
	if bodyStr, ok := args["body"].(string); ok && bodyStr != "" {
		body = strings.NewReader(bodyStr)
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Agentry/1.0")
	if headers, ok := args["headers"].(map[string]any); ok {
		for key, value := range headers {
			if v, ok := value.(string); ok {
				req.Header.Set(key, v)
			}
		}
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		if bodyStr, ok := args["body"].(string); ok {
			bodyStr = strings.TrimSpace(bodyStr)
			if strings.HasPrefix(bodyStr, "{") || strings.HasPrefix(bodyStr, "[") {
				req.Header.Set("Content-Type", "application/json")
			}
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	result := map[string]any{
		"url":            urlStr,
		"method":         method,
		"status_code":    resp.StatusCode,
		"status":         resp.Status,
		"headers":        resp.Header,
		"body":           string(respBody),
		"content_type":   resp.Header.Get("Content-Type"),
		"content_length": len(respBody),
		"duration_ms":    time.Since(start).Milliseconds(),
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var jsonData any
		if err := json.Unmarshal(respBody, &jsonData); err == nil {
			result["json"] = jsonData
		}
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
