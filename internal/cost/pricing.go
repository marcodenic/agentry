package cost

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ModelPricing holds the pricing information for a specific model
type ModelPricing struct {
	InputPrice  float64 // Price per 1M tokens for input
	OutputPrice float64 // Price per 1M tokens for output
}

// PricingTable holds all model pricing information
type PricingTable struct {
	mu     sync.RWMutex
	prices map[string]ModelPricing
}

// NewPricingTable creates a new pricing table with cached prices
func NewPricingTable() *PricingTable {
	pt := &PricingTable{
		prices: make(map[string]ModelPricing),
	}
	pt.loadPrices()
	return pt
}

// loadPrices loads pricing data from cached file or falls back to defaults
func (pt *PricingTable) loadPrices() {
	// Try to load from cached file first
	if pt.loadFromCache() {
		return
	}

	// If no cached file, load minimal defaults
	pt.loadDefaultPrices()
}

// loadFromCache loads pricing data from the cached JSON file
func (pt *PricingTable) loadFromCache() bool {
	cacheFile := pt.getCacheFilePath()

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return false
	}

	var apiData map[string]interface{}
	if err := json.Unmarshal(data, &apiData); err != nil {
		return false
	}

	// Parse the cached API data
	pt.parseAPIData(apiData)
	return true
}

// loadDefaultPrices loads minimal fallback pricing data
func (pt *PricingTable) loadDefaultPrices() {
	// Only minimal fallback prices for when API is unavailable
	defaultPrices := map[string]ModelPricing{
		// Basic OpenAI models
		"gpt-4":         {InputPrice: 30.0, OutputPrice: 60.0},
		"gpt-4o":        {InputPrice: 2.5, OutputPrice: 10.0},
		"gpt-4o-mini":   {InputPrice: 0.15, OutputPrice: 0.6},
		"gpt-3.5-turbo": {InputPrice: 0.5, OutputPrice: 1.5},

		// Basic Claude models
		"claude-3-opus":     {InputPrice: 15.0, OutputPrice: 75.0},
		"claude-3-sonnet":   {InputPrice: 3.0, OutputPrice: 15.0},
		"claude-3-haiku":    {InputPrice: 0.25, OutputPrice: 1.25},
		"claude-3-5-sonnet": {InputPrice: 3.0, OutputPrice: 15.0},
		"claude-3-5-haiku":  {InputPrice: 1.0, OutputPrice: 5.0},
	}

	for model, pricing := range defaultPrices {
		pt.prices[model] = pricing
	}
}

// parseAPIData extracts pricing information from the models.dev API response
func (pt *PricingTable) parseAPIData(apiData map[string]interface{}) {
	for providerID, providerData := range apiData {
		provider, ok := providerData.(map[string]interface{})
		if !ok {
			continue
		}

		models, ok := provider["models"].(map[string]interface{})
		if !ok {
			continue
		}

		for modelID, modelData := range models {
			model, ok := modelData.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract cost information
			cost, ok := model["cost"].(map[string]interface{})
			if !ok {
				continue
			}

			inputPrice, inputOk := cost["input"].(float64)
			outputPrice, outputOk := cost["output"].(float64)

			if inputOk && outputOk {
				// Store with both provider prefix and clean model name
				fullModelName := fmt.Sprintf("%s/%s", providerID, modelID)
				pt.prices[fullModelName] = ModelPricing{InputPrice: inputPrice, OutputPrice: outputPrice}
				pt.prices[modelID] = ModelPricing{InputPrice: inputPrice, OutputPrice: outputPrice}
			}
		}
	}
}

// getCacheFilePath returns the path to the cached pricing file
func (pt *PricingTable) getCacheFilePath() string {
	return filepath.Join("internal", "cost", "data", "models_pricing.json")
}

// GetPricing returns the pricing for a given model
func (pt *PricingTable) GetPricing(model string) (ModelPricing, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	// Try exact match first
	if pricing, ok := pt.prices[model]; ok {
		return pricing, true
	}

	// Try partial matching for versioned models - use deterministic order
	modelLower := strings.ToLower(model)
	
	// Collect all potential matches first
	var matches []struct {
		priceModel string
		pricing    ModelPricing
		matchType  int // 0 = exact contains, 1 = partial contains
	}
	
	for priceModel, pricing := range pt.prices {
		priceModelLower := strings.ToLower(priceModel)
		
		// Prioritize exact substring matches
		if strings.Contains(modelLower, priceModelLower) {
			matches = append(matches, struct {
				priceModel string
				pricing    ModelPricing
				matchType  int
			}{priceModel, pricing, 0})
		} else if strings.Contains(priceModelLower, modelLower) {
			matches = append(matches, struct {
				priceModel string
				pricing    ModelPricing
				matchType  int
			}{priceModel, pricing, 1})
		}
	}
	
	// If we have matches, return the first one with the highest priority
	// This ensures deterministic behavior
	if len(matches) > 0 {
		// Sort by match type (exact matches first), then by model name for determinism
		bestMatch := matches[0]
		for _, match := range matches[1:] {
			if match.matchType < bestMatch.matchType ||
				(match.matchType == bestMatch.matchType && match.priceModel < bestMatch.priceModel) {
				bestMatch = match
			}
		}
		return bestMatch.pricing, true
	}

	return ModelPricing{}, false
}

// CalculateCost calculates the cost for input and output tokens
func (pt *PricingTable) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing, found := pt.GetPricing(model)
	if !found {
		// Fallback to a reasonable default if model not found
		pricing = ModelPricing{InputPrice: 1.0, OutputPrice: 3.0}
	}

	// Convert tokens to millions and calculate cost
	inputCost := float64(inputTokens) * pricing.InputPrice / 1000000.0
	outputCost := float64(outputTokens) * pricing.OutputPrice / 1000000.0

	return inputCost + outputCost
}

// RefreshFromAPI downloads fresh pricing data from the models.dev API and caches it
func (pt *PricingTable) RefreshFromAPI() error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://models.dev/api.json")
	if err != nil {
		return fmt.Errorf("failed to fetch pricing data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %w", err)
	}

	var apiData map[string]interface{}
	if err := json.Unmarshal(body, &apiData); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	// Save the raw API data to cache file
	cacheFile := pt.getCacheFilePath()
	if err := os.WriteFile(cacheFile, body, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Parse and update pricing data
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.prices = make(map[string]ModelPricing) // Clear existing prices
	pt.parseAPIData(apiData)

	return nil
}

// UpdateFromAPI updates pricing from the models.dev API (deprecated, use RefreshFromAPI)
func (pt *PricingTable) UpdateFromAPI() error {
	return pt.RefreshFromAPI()
}

// SetCustomPricing allows setting custom pricing for a model
func (pt *PricingTable) SetCustomPricing(model string, inputPrice, outputPrice float64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.prices[model] = ModelPricing{InputPrice: inputPrice, OutputPrice: outputPrice}
}

// ListModels returns all models with pricing information
func (pt *PricingTable) ListModels() map[string]ModelPricing {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	result := make(map[string]ModelPricing)
	for model, pricing := range pt.prices {
		result[model] = pricing
	}
	return result
}

// GetCachedDataAge returns how old the cached data is, or an error if no cache exists
func (pt *PricingTable) GetCachedDataAge() (time.Duration, error) {
	cacheFile := pt.getCacheFilePath()
	info, err := os.ReadFile(cacheFile)
	if err != nil {
		return 0, fmt.Errorf("no cached data found")
	}

	// Get file modification time would be better, but for now return 0 if file exists
	if len(info) > 0 {
		return 0, nil // File exists
	}
	return 0, fmt.Errorf("no cached data found")
}
