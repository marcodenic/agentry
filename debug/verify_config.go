package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/config"
)

func main() {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting working directory: %v\n", err)
		return
	}

	// Load the smart config
	configPath := filepath.Join(wd, "config", "smart-config.yaml")
	cfg, err := config.FromFile(configPath)
	if err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		return
	}

	fmt.Printf("📋 Config loaded successfully\n")
	fmt.Printf("🔧 Number of models: %d\n", len(cfg.Models))

	for i, m := range cfg.Models {
		fmt.Printf("📊 Model %d:\n", i)
		fmt.Printf("  - Name: %s\n", m.Name)
		fmt.Printf("  - Provider: %s\n", m.Provider)
		if m.Options != nil {
			fmt.Printf("  - Options:\n")
			for k, v := range m.Options {
				fmt.Printf("    - %s: %s\n", k, v)
			}
		}

		// Simulate the model name construction logic from buildAgent
		var modelName string
		if m.Options != nil && m.Options["model"] != "" {
			modelName = fmt.Sprintf("%s-%s", m.Provider, m.Options["model"])
		} else {
			modelName = m.Name
		}
		fmt.Printf("  - Constructed ModelName: %s\n", modelName)
		fmt.Printf("\n")
	}
}
