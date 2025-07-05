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
	InputPrice    float64 // Price per 1M tokens for input
	OutputPrice   float64 // Price per 1M tokens for output
	ContextLimit  int     // Maximum context window size in tokens
	OutputLimit   int     // Maximum output tokens
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
	if err := pt.loadPrices(); err != nil {
		// Log error but don't panic - allow the system to continue with empty pricing
		// This will result in zero costs for unknown models
		return pt
	}
	return pt
}

// loadPrices loads pricing data from cached file
func (pt *PricingTable) loadPrices() error {
	// Only load from cache - no hardcoded defaults
	if !pt.loadFromCache() {
		return fmt.Errorf("pricing cache file is missing or corrupted. Please run 'agentry refresh-pricing' to download fresh pricing data")
	}
	return nil
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



// parseAPIData extracts pricing information from the models.dev API response
func (pt *PricingTable) parseAPIData(apiData map[string]interface{}) {
	// First pass: collect all models with provider prefixes
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

			// Extract context limits
			var contextLimit, outputLimit int
			if limit, ok := model["limit"].(map[string]interface{}); ok {
				if ctx, ok := limit["context"].(float64); ok {
					contextLimit = int(ctx)
				}
				if out, ok := limit["output"].(float64); ok {
					outputLimit = int(out)
				}
			}

			if inputOk && outputOk {
				// Store with provider prefix
				fullModelName := fmt.Sprintf("%s/%s", providerID, modelID)
				pt.prices[fullModelName] = ModelPricing{
					InputPrice:   inputPrice, 
					OutputPrice:  outputPrice,
					ContextLimit: contextLimit,
					OutputLimit:  outputLimit,
				}
			}
		}
	}
	
	// Only store provider/model format - no fallback to plain model names
}

// getCacheFilePath returns the path to the cached pricing file
func (pt *PricingTable) getCacheFilePath() string {
	// Try to find the module root by looking for go.mod
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path from current directory
		return filepath.Join("internal", "cost", "data", "models_pricing.json")
	}
	
	// Look for go.mod starting from current directory and going up
	dir := cwd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, this is the module root
			return filepath.Join(dir, "internal", "cost", "data", "models_pricing.json")
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, fallback to relative path
			break
		}
		dir = parent
	}
	
	// Fallback to relative path
	return filepath.Join("internal", "cost", "data", "models_pricing.json")
}

// GetPricing returns the pricing for a given model
func (pt *PricingTable) GetPricing(model string) (ModelPricing, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	// Only exact match - no fuzzy matching
	if pricing, ok := pt.prices[model]; ok {
		return pricing, true
	}

	return ModelPricing{}, false
}

// GetPricingByProvider returns the pricing for a given provider and model
func (pt *PricingTable) GetPricingByProvider(provider, model string) (ModelPricing, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	// Only try provider/model format - no fallback to plain model names
	providerModel := fmt.Sprintf("%s/%s", provider, model)
	if pricing, ok := pt.prices[providerModel]; ok {
		return pricing, true
	}

	return ModelPricing{}, false
}

// GetPricingByModelName handles provider-model format names like "openai-gpt-4" or "anthropic-claude-instant"
func (pt *PricingTable) GetPricingByModelName(modelName string) (ModelPricing, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	// Try exact match first (for provider/model format like "openai/gpt-4")
	if pricing, ok := pt.prices[modelName]; ok {
		return pricing, true
	}

	// Try to parse provider-model format (like "openai-gpt-4" -> "openai/gpt-4")
	parts := strings.Split(modelName, "-")
	if len(parts) >= 2 {
		provider := parts[0]
		model := strings.Join(parts[1:], "-")
		
		// Try provider/model format
		providerModel := fmt.Sprintf("%s/%s", provider, model)
		if pricing, ok := pt.prices[providerModel]; ok {
			return pricing, true
		}
	}

	// Fuzzy matching for models with suffixes like -latest, -beta, -preview, etc.
	return pt.findBestMatch(modelName)
}

// findBestMatch attempts to find the best pricing match for a model name
// This handles cases like "claude-3-7-sonnet-latest" matching "claude-3-7-sonnet-20250219"
func (pt *PricingTable) findBestMatch(modelName string) (ModelPricing, bool) {
	// Common suffixes that don't affect pricing
	suffixes := []string{"-latest", "-beta", "-preview", "-alpha", "-rc", "-stable"}
	
	// Try removing each suffix
	for _, suffix := range suffixes {
		if strings.HasSuffix(modelName, suffix) {
			baseModel := strings.TrimSuffix(modelName, suffix)
			if pricing, ok := pt.prices[baseModel]; ok {
				return pricing, true
			}
		}
	}
	
	// For provider/model format, also try suffix removal and fuzzy matching
	if strings.Contains(modelName, "/") {
		parts := strings.Split(modelName, "/")
		if len(parts) == 2 {
			provider := parts[0]
			model := parts[1]
			
			// Try removing suffixes from the model part
			for _, suffix := range suffixes {
				if strings.HasSuffix(model, suffix) {
					baseModel := strings.TrimSuffix(model, suffix)
					candidateKey := fmt.Sprintf("%s/%s", provider, baseModel)
					if pricing, ok := pt.prices[candidateKey]; ok {
						return pricing, true
					}
				}
			}
			
			// Try fuzzy matching against all models with the same provider
			baseModel := model
			for _, suffix := range suffixes {
				if strings.HasSuffix(baseModel, suffix) {
					baseModel = strings.TrimSuffix(baseModel, suffix)
				}
			}
			
			// Find best match by checking if any pricing key starts with provider/baseModel
			bestMatch := ""
			for pricingKey := range pt.prices {
				if strings.HasPrefix(pricingKey, provider+"/") {
					pricingModel := strings.TrimPrefix(pricingKey, provider+"/")
					// Check if this pricing model starts with our base model
					if strings.HasPrefix(pricingModel, baseModel) {
						// Prefer shorter matches (fewer extra characters)
						if bestMatch == "" || len(pricingModel) < len(strings.TrimPrefix(bestMatch, provider+"/")) {
							bestMatch = pricingKey
						}
					}
				}
			}
			
			if bestMatch != "" {
				if pricing, ok := pt.prices[bestMatch]; ok {
					return pricing, true
				}
			}
		}
	}
	
	// Try progressive shortening for versioned models
	// e.g., "claude-3-7-sonnet" -> "claude-3-7" -> "claude-3"
	if strings.Contains(modelName, "/") {
		parts := strings.Split(modelName, "/")
		if len(parts) == 2 {
			provider := parts[0]
			model := parts[1]
			
			modelParts := strings.Split(model, "-")
			// Try progressively shorter versions
			for i := len(modelParts) - 1; i >= 2; i-- {
				shorterModel := strings.Join(modelParts[:i], "-")
				candidateKey := fmt.Sprintf("%s/%s", provider, shorterModel)
				if pricing, ok := pt.prices[candidateKey]; ok {
					return pricing, true
				}
			}
		}
	}
	
	return ModelPricing{}, false
}

// CalculateCost calculates the cost for input and output tokens
func (pt *PricingTable) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing, found := pt.GetPricingByModelName(model)
	if !found {
		// Return zero cost if model not found - no hardcoded fallbacks
		return 0.0
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

// GetContextLimit returns the context window limit for a given model
func (pt *PricingTable) GetContextLimit(modelName string) int {
	pricing, found := pt.GetPricingByModelName(modelName)
	if !found || pricing.ContextLimit == 0 {
		// Fallback to reasonable defaults if no pricing data found
		if strings.Contains(strings.ToLower(modelName), "gpt-4") {
			return 128000
		}
		if strings.Contains(strings.ToLower(modelName), "claude") {
			return 200000
		}
		return 8000 // Conservative default
	}
	return pricing.ContextLimit
}
