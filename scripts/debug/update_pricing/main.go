package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run update_pricing.go [update-from-api|list-models|test-pricing]")
		return
	}

	command := os.Args[1]
	pricing := cost.NewPricingTable()

	switch command {
	case "update-from-api":
		fmt.Println("Fetching pricing data from models.dev API...")
		if err := fetchAndDisplayAPIData(); err != nil {
			fmt.Printf("Error fetching API data: %v\n", err)
			fmt.Println("The API structure might need to be implemented when we see the actual format.")
		}
	case "list-models":
		fmt.Println("Available models and their pricing:")
		models := pricing.ListModels()
		for model, modelPricing := range models {
			fmt.Printf("%-30s: input=$%.2f/MTok, output=$%.2f/MTok\n",
				model, modelPricing.InputPrice, modelPricing.OutputPrice)
		}
	case "test-pricing":
		fmt.Println("Testing pricing for common usage patterns...")
		testModels := []string{"gpt-4", "gpt-4o", "gpt-4.1", "claude-3-opus", "claude-3-sonnet"}
		testTokens := []int{1000, 10000, 100000}

		for _, model := range testModels {
			fmt.Printf("\n%s pricing:\n", model)
			for _, tokens := range testTokens {
				cost := pricing.CalculateCost(model, tokens, tokens)
				fmt.Printf("  %d in + %d out tokens: $%.6f\n", tokens, tokens, cost)
			}
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: update-from-api, list-models, test-pricing")
	}
}

func fetchAndDisplayAPIData() error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://models.dev/api.json")
	if err != nil {
		return fmt.Errorf("failed to fetch API data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Pretty print the JSON to see the structure
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	fmt.Println("API Response Structure:")
	fmt.Println(string(prettyJSON))

	return nil
}
