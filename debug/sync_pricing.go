package main

import (
	"fmt"
	"log"
	"os"

	"github.com/marcodenic/agentry/internal/cost"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run sync_pricing.go <update|check>")
		os.Exit(1)
	}

	command := os.Args[1]
	pricing := cost.NewPricingTable()

	switch command {
	case "check":
		fmt.Println("Checking current pricing...")
		models := pricing.ListModels()
		for modelName, modelPricing := range models {
			fmt.Printf("%-30s : input=$%.2f/MTok, output=$%.2f/MTok\n",
				modelName, modelPricing.InputPrice, modelPricing.OutputPrice)
		}
	case "update":
		fmt.Println("Updating pricing from models.dev API...")
		err := pricing.UpdateFromAPI()
		if err != nil {
			log.Printf("Warning: Could not update pricing from API: %v", err)
			fmt.Println("Using hardcoded pricing data instead.")
		} else {
			fmt.Println("Pricing updated successfully from API!")
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: go run sync_pricing.go <update|check>")
		os.Exit(1)
	}
}
