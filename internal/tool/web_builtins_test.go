package tool

import (
	"context"
	"encoding/json"
	"testing"
)

func TestWebTools(t *testing.T) {
	registry := DefaultRegistry()

	t.Run("web_search", func(t *testing.T) {
		tool, exists := registry.Use("web_search")
		if !exists {
			t.Fatal("web_search tool not found in registry")
		}

		args := map[string]any{
			"query":       "golang testing",
			"provider":    "duckduckgo",
			"max_results": 3.0,
		}

		result, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Fatalf("web_search failed: %v", err)
		}

		var searchResult map[string]any
		if err := json.Unmarshal([]byte(result), &searchResult); err != nil {
			t.Fatalf("Failed to parse web_search result: %v", err)
		}

		if searchResult["query"] != "golang testing" {
			t.Error("Query not preserved in result")
		}

		if searchResult["provider"] != "duckduckgo" {
			t.Error("Provider not preserved in result")
		}

		if results, ok := searchResult["results"].([]any); !ok {
			t.Error("Expected search results array")
		} else if len(results) == 0 {
			// DuckDuckGo API might return empty results - this is acceptable
			t.Log("No search results returned (DuckDuckGo API behavior)")
		}
	})

	t.Run("read_webpage", func(t *testing.T) {
		tool, exists := registry.Use("read_webpage")
		if !exists {
			t.Fatal("read_webpage tool not found in registry")
		}

		// Test with a reliable site
		args := map[string]any{
			"url":        "https://httpbin.org/html",
			"extract":    "title",
			"max_length": 1000.0,
		}

		result, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Logf("read_webpage failed (may be network issue): %v", err)
			t.Skip("Skipping due to network dependency")
		}

		var webResult map[string]any
		if err := json.Unmarshal([]byte(result), &webResult); err != nil {
			t.Fatalf("Failed to parse read_webpage result: %v", err)
		}

		if webResult["type"] != "title" {
			t.Error("Expected type=title")
		}

		if webResult["url"] != "https://httpbin.org/html" {
			t.Error("URL not preserved in result")
		}
	})

	t.Run("api_request", func(t *testing.T) {
		tool, exists := registry.Use("api_request")
		if !exists {
			t.Fatal("api_request tool not found in registry")
		}

		// Test with httpbin for reliable testing
		args := map[string]any{
			"url":    "https://httpbin.org/json",
			"method": "GET",
			"headers": map[string]any{
				"Accept": "application/json",
			},
			"timeout": 30.0,
		}

		result, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Logf("api_request failed (may be network issue): %v", err)
			t.Skip("Skipping due to network dependency")
		}

		var apiResult map[string]any
		if err := json.Unmarshal([]byte(result), &apiResult); err != nil {
			t.Fatalf("Failed to parse api_request result: %v", err)
		}

		if apiResult["method"] != "GET" {
			t.Error("Method not preserved in result")
		}

		if apiResult["status_code"] != 200.0 {
			t.Errorf("Expected status code 200, got %v", apiResult["status_code"])
		}

		if body, ok := apiResult["body"].(string); !ok || body == "" {
			t.Error("Expected response body")
		}
	})
}

func TestWebToolsInRegistry(t *testing.T) {
	registry := DefaultRegistry()

	expectedTools := []string{
		"web_search",
		"read_webpage",
		"api_request",
		"download_file",
	}

	for _, toolName := range expectedTools {
		tool, exists := registry.Use(toolName)
		if !exists {
			t.Errorf("Tool %s not found in registry", toolName)
			continue
		}

		if tool.Name() != toolName {
			t.Errorf("Tool name mismatch: expected %s, got %s", toolName, tool.Name())
		}

		schema := tool.JSONSchema()
		if schema == nil {
			t.Errorf("Tool %s has no schema", toolName)
		}
	}
}
