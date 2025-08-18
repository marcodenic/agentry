package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Change to project root
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}

	fmt.Println("Testing direct JSON parsing...")

	// Read the pricing file directly
	data, err := os.ReadFile("internal/cost/data/models_pricing.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var apiData map[string]interface{}
	if err := json.Unmarshal(data, &apiData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	fmt.Printf("Successfully parsed JSON with %d top-level keys\n", len(apiData))
	
	// Look for openai provider
	openaiData, ok := apiData["openai"]
	if !ok {
		fmt.Println("❌ No 'openai' key found")
		fmt.Println("Available keys:")
		for key := range apiData {
			fmt.Printf("  - %s\n", key)
		}
		return
	}
	
	fmt.Println("✅ Found openai data")
	
	// Check if it's the right structure
	openaiProvider, ok := openaiData.(map[string]interface{})
	if !ok {
		fmt.Println("❌ openai data is not a map")
		return
	}
	
	// Look for models
	models, ok := openaiProvider["models"]
	if !ok {
		fmt.Println("❌ No 'models' key in openai data")
		fmt.Println("Available keys in openai:")
		for key := range openaiProvider {
			fmt.Printf("  - %s\n", key)
		}
		return
	}
	
	fmt.Println("✅ Found models in openai data")
	
	// Check models structure
	modelsMap, ok := models.(map[string]interface{})
	if !ok {
		fmt.Println("❌ models is not a map")
		return
	}
	
	fmt.Printf("Found %d models in openai\n", len(modelsMap))
	
	// Look for gpt-4.1-nano specifically
	gpt41nano, ok := modelsMap["gpt-4.1-nano"]
	if !ok {
		fmt.Println("❌ gpt-4.1-nano not found in models")
		fmt.Println("Available models:")
		for modelName := range modelsMap {
			fmt.Printf("  - %s\n", modelName)
		}
		return
	}
	
	fmt.Println("✅ Found gpt-4.1-nano in models")
	
	// Check the model structure
	modelData, ok := gpt41nano.(map[string]interface{})
	if !ok {
		fmt.Println("❌ gpt-4.1-nano data is not a map")
		return
	}
	
	// Look for cost data
	cost, ok := modelData["cost"]
	if !ok {
		fmt.Println("❌ No 'cost' key in gpt-4.1-nano")
		fmt.Println("Available keys in gpt-4.1-nano:")
		for key := range modelData {
			fmt.Printf("  - %s\n", key)
		}
		return
	}
	
	fmt.Println("✅ Found cost data in gpt-4.1-nano")
	
	// Check cost structure
	costData, ok := cost.(map[string]interface{})
	if !ok {
		fmt.Println("❌ cost data is not a map")
		return
	}
	
	// Get input and output prices
	inputPrice, inputOk := costData["input"].(float64)
	outputPrice, outputOk := costData["output"].(float64)
	
	if !inputOk || !outputOk {
		fmt.Println("❌ Input or output price not found or not float64")
		fmt.Printf("input: %v (type: %T)\n", costData["input"], costData["input"])
		fmt.Printf("output: %v (type: %T)\n", costData["output"], costData["output"])
		return
	}
	
	fmt.Printf("✅ Successfully extracted pricing:\n")
	fmt.Printf("  Input price: $%.3f per 1M tokens\n", inputPrice)
	fmt.Printf("  Output price: $%.3f per 1M tokens\n", outputPrice)
	
	// Calculate cost for 1053 tokens
	cost1053 := float64(1053) * inputPrice / 1000000.0
	fmt.Printf("  Cost for 1053 input tokens: $%.9f\n", cost1053)
}
