package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	// Add web and network capability tools
	addWebTools()
}

func addWebTools() {
	// web_search - Search the web using multiple providers
	builtinMap["web_search"] = builtinSpec{
		Desc: "Search the web using various search engines",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query",
				},
				"provider": map[string]any{
					"type":        "string",
					"description": "Search provider (duckduckgo, bing, google). Default: duckduckgo",
					"enum":        []string{"duckduckgo", "bing", "google"},
					"default":     "duckduckgo",
				},
				"max_results": map[string]any{
					"type":        "integer",
					"description": "Maximum number of results to return (1-20, default: 10)",
					"minimum":     1,
					"maximum":     20,
					"default":     10,
				},
			},
			"required": []string{"query"},
			"example": map[string]any{
				"query":       "Go programming best practices",
				"provider":    "duckduckgo",
				"max_results": 5,
			},
		},
		Exec: webSearchExec,
	}

	// read_webpage - Extract content from web pages
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

	// api_request - Generic HTTP/REST API calls
	builtinMap["api_request"] = builtinSpec{
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
					"type":        "object",
					"description": "HTTP headers to send",
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

	// download_file - Download files from URLs
	builtinMap["download_file"] = builtinSpec{
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

// webSearchExec implements the web_search tool
func webSearchExec(ctx context.Context, args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", errors.New("missing query")
	}

	provider := "duckduckgo"
	if p, ok := args["provider"].(string); ok {
		provider = p
	}

	maxResults := 10
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	var results []map[string]any
	var err error

	switch provider {
	case "duckduckgo":
		results, err = searchDuckDuckGo(ctx, query, maxResults)
	case "bing":
		results, err = searchBing(ctx, query, maxResults)
	case "google":
		results, err = searchGoogle(ctx, query, maxResults)
	default:
		return "", fmt.Errorf("unsupported search provider: %s", provider)
	}

	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	response := map[string]any{
		"query":     query,
		"provider":  provider,
		"count":     len(results),
		"results":   results,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonResult, _ := json.Marshal(response)
	return string(jsonResult), nil
}

// searchDuckDuckGo performs search using DuckDuckGo
func searchDuckDuckGo(ctx context.Context, query string, maxResults int) ([]map[string]any, error) {
	// DuckDuckGo instant answer API
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query))
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Agentry/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var ddgResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&ddgResp); err != nil {
		return nil, err
	}
	
	var results []map[string]any
	
	// Extract instant answer if available
	if abstract, ok := ddgResp["Abstract"].(string); ok && abstract != "" {
		result := map[string]any{
			"title":       ddgResp["Heading"],
			"url":         ddgResp["AbstractURL"],
			"description": abstract,
			"type":        "instant_answer",
		}
		results = append(results, result)
	}
	
	// Extract related topics
	if topics, ok := ddgResp["RelatedTopics"].([]any); ok {
		for i, topic := range topics {
			if i >= maxResults-len(results) {
				break
			}
			if topicMap, ok := topic.(map[string]any); ok {
				if text, ok := topicMap["Text"].(string); ok && text != "" {
					result := map[string]any{
						"title":       extractTitle(text),
						"url":         topicMap["FirstURL"],
						"description": text,
						"type":        "related_topic",
					}
					results = append(results, result)
				}
			}
		}
	}
	
	return results, nil
}

// searchBing performs search using Bing (placeholder - would need API key)
func searchBing(ctx context.Context, query string, maxResults int) ([]map[string]any, error) {
	// Note: This would require Bing Search API subscription
	// For now, return a placeholder result
	return []map[string]any{
		{
			"title":       "Bing Search API Required",
			"url":         "https://www.microsoft.com/en-us/bing/apis/bing-web-search-api",
			"description": "Bing search requires API key configuration",
			"type":        "info",
		},
	}, nil
}

// searchGoogle performs search using Google (placeholder - would need API key)
func searchGoogle(ctx context.Context, query string, maxResults int) ([]map[string]any, error) {
	// Note: This would require Google Custom Search API
	// For now, return a placeholder result
	return []map[string]any{
		{
			"title":       "Google Search API Required",
			"url":         "https://developers.google.com/custom-search/v1/overview",
			"description": "Google search requires Custom Search API key configuration",
			"type":        "info",
		},
	}, nil
}

// extractTitle extracts a title from text
func extractTitle(text string) string {
	words := strings.Fields(text)
	if len(words) > 8 {
		return strings.Join(words[:8], " ") + "..."
	}
	return text
}

// readWebpageExec implements the read_webpage tool
func readWebpageExec(ctx context.Context, args map[string]any) (string, error) {
	urlStr, _ := args["url"].(string)
	if urlStr == "" {
		return "", errors.New("missing url")
	}

	extract := "text"
	if e, ok := args["extract"].(string); ok {
		extract = e
	}

	maxLength := 10000
	if ml, ok := args["max_length"].(float64); ok {
		maxLength = int(ml)
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("only HTTP/HTTPS URLs are supported")
	}

	// Fetch the webpage
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
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

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse and extract content
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
		if len(links) > 100 { // Limit number of links
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
		if len(links) > 50 { // Limit for "all" mode
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

// extractTextSimple extracts plain text from HTML using regex
func extractTextSimple(html string) string {
	// Remove script and style tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")
	
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")
	
	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")
	
	// Clean up whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

// extractTitleSimple extracts the page title using regex
func extractTitleSimple(html string) string {
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractLinksSimple extracts links using regex
func extractLinksSimple(html string, baseURL *url.URL) []map[string]any {
	var links []map[string]any
	
	linkRegex := regexp.MustCompile(`(?i)<a[^>]*href\s*=\s*["']([^"']*)["'][^>]*>(.*?)</a>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)
	
	for _, match := range matches {
		if len(match) > 2 {
			href := match[1]
			text := extractTextSimple(match[2])
			
			if href != "" {
				// Resolve relative URLs
				if linkURL, err := baseURL.Parse(href); err == nil {
					if text == "" {
						text = href
					}
					links = append(links, map[string]any{
						"url":  linkURL.String(),
						"text": text,
					})
				}
			}
		}
	}
	
	return links
}

// apiRequestExec implements the api_request tool
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

	// Create request
	var body io.Reader
	if bodyStr, ok := args["body"].(string); ok && bodyStr != "" {
		body = strings.NewReader(bodyStr)
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Agentry/1.0")
	if headers, ok := args["headers"].(map[string]any); ok {
		for key, value := range headers {
			if valueStr, ok := value.(string); ok {
				req.Header.Set(key, valueStr)
			}
		}
	}

	// Set content type for POST/PUT requests with body
	if body != nil && req.Header.Get("Content-Type") == "" {
		// Try to detect if it's JSON
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

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	result := map[string]any{
		"url":           urlStr,
		"method":        method,
		"status_code":   resp.StatusCode,
		"status":        resp.Status,
		"headers":       resp.Header,
		"body":          string(respBody),
		"content_type":  resp.Header.Get("Content-Type"),
		"content_length": len(respBody),
		"duration_ms":   time.Since(start).Milliseconds(),
	}

	// Try to parse JSON response
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var jsonData any
		if err := json.Unmarshal(respBody, &jsonData); err == nil {
			result["json"] = jsonData
		}
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

// downloadFileExec implements the download_file tool
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
	}

	path = absPath(path)

	// Check if file exists
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("file %s already exists (use overwrite=true to replace)", path)
		}
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Download file
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
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

	// Check content length if available
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			sizeMB := float64(size) / (1024 * 1024)
			if sizeMB > maxSizeMB {
				return "", fmt.Errorf("file size %.1fMB exceeds limit %.1fMB", sizeMB, maxSizeMB)
			}
		}
	}

	// Create temp file
	tempPath := path + ".download"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy with size limit
	maxBytes := int64(maxSizeMB * 1024 * 1024)
	written, err := io.CopyN(tempFile, resp.Body, maxBytes)
	tempFile.Close()

	if err != nil && err != io.EOF {
		os.Remove(tempPath)
		return "", fmt.Errorf("download failed: %w", err)
	}

	// Check if we hit the size limit
	if written == maxBytes {
		os.Remove(tempPath)
		return "", fmt.Errorf("file exceeds size limit %.1fMB", maxSizeMB)
	}

	// Atomic move
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
