package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func init() {
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
}

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

	var (
		results []map[string]any
		err     error
	)

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

func searchDuckDuckGo(ctx context.Context, query string, maxResults int) ([]map[string]any, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
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

	if abstract, ok := ddgResp["Abstract"].(string); ok && abstract != "" {
		result := map[string]any{
			"title":       ddgResp["Heading"],
			"url":         ddgResp["AbstractURL"],
			"description": abstract,
			"type":        "instant_answer",
		}
		results = append(results, result)
	}

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

func searchBing(ctx context.Context, query string, maxResults int) ([]map[string]any, error) {
	return []map[string]any{{
		"title":       "Bing Search API Required",
		"url":         "https://www.microsoft.com/en-us/bing/apis/bing-web-search-api",
		"description": "Bing search requires API key configuration",
		"type":        "info",
	}}, nil
}

func searchGoogle(ctx context.Context, query string, maxResults int) ([]map[string]any, error) {
	return []map[string]any{{
		"title":       "Google Search API Required",
		"url":         "https://developers.google.com/custom-search/v1/overview",
		"description": "Google search requires Custom Search API key configuration",
		"type":        "info",
	}}, nil
}
