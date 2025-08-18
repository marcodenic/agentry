package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	builtinMap["read_webpage"] = builtinSpec{
		Desc: "Read and extract content from web pages",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "URL to read",
				},
				"extract": map[string]any{
					"type":        "string",
					"description": "What to extract (text, title, links, all). Default: text",
					"enum":        []string{"text", "title", "links", "all"},
					"default":     "text",
				},
				"max_length": map[string]any{
					"type":        "integer",
					"description": "Maximum content length to return (default: 10000)",
					"minimum":     100,
					"maximum":     50000,
					"default":     10000,
				},
			},
			"required": []string{"url"},
			"example": map[string]any{
				"url":        "https://golang.org/doc/",
				"extract":    "text",
				"max_length": 5000,
			},
		},
		Exec: readWebpageExec,
	}
}

func readWebpageExec(ctx context.Context, args map[string]any) (string, error) {
	urlStr, _ := args["url"].(string)
	if urlStr == "" {
		return "", errors.New("missing url")
	}

	extract := "text"
	if e, ok := args["extract"].(string); ok {
		extract = e
	}

	maxLength, _ := getIntArg(args, "max_length", 10000)

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("only HTTP/HTTPS URLs are supported")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Agentry/1.0 (Web Content Reader)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]any

	switch extract {
	case "text":
		text := extractTextSimple(string(body))
		if len(text) > maxLength {
			text = text[:maxLength] + "..."
		}
		result = map[string]any{
			"url":     urlStr,
			"type":    "text",
			"content": text,
			"length":  len(text),
		}
	case "title":
		title := extractTitleSimple(string(body))
		result = map[string]any{
			"url":   urlStr,
			"type":  "title",
			"title": title,
		}
	case "links":
		links := extractLinksSimple(string(body), parsedURL)
		if len(links) > 100 {
			links = links[:100]
		}
		result = map[string]any{
			"url":   urlStr,
			"type":  "links",
			"links": links,
			"count": len(links),
		}
	case "all":
		text := extractTextSimple(string(body))
		if len(text) > maxLength {
			text = text[:maxLength] + "..."
		}
		title := extractTitleSimple(string(body))
		links := extractLinksSimple(string(body), parsedURL)
		if len(links) > 50 {
			links = links[:50]
		}
		result = map[string]any{
			"url":     urlStr,
			"type":    "all",
			"title":   title,
			"content": text,
			"links":   links,
			"length":  len(text),
		}
	default:
		return "", fmt.Errorf("unsupported extract type: %s", extract)
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
