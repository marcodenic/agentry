package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RoleConfig represents a role configuration from YAML
type RoleConfig struct {
	Name   string   `yaml:"name"`
	Prompt string   `yaml:"prompt"`
	Tools  []string `yaml:"tools,omitempty"`
}

// loadRolePrompt loads a role-specific prompt from the templates/roles directory
func loadRolePrompt(roleName string) (string, error) {
	// Try to find the templates/roles directory by searching up from the current directory
	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	fmt.Printf("Starting search from: %s\n", workDir)
	
	for dir := workDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		templatesDir := filepath.Join(dir, "templates", "roles")
		fmt.Printf("Checking: %s\n", templatesDir)
		if _, err := os.Stat(templatesDir); err == nil {
			roleFile := filepath.Join(templatesDir, roleName+".yaml")
			fmt.Printf("Looking for role file: %s\n", roleFile)
			if _, err := os.Stat(roleFile); err == nil {
				data, err := os.ReadFile(roleFile)
				if err != nil {
					return "", err
				}
				
				var config RoleConfig
				if err := yaml.Unmarshal(data, &config); err != nil {
					return "", err
				}
				
				fmt.Printf("Loaded prompt for %s: %s\n", roleName, config.Prompt[:50]+"...")
				return config.Prompt, nil
			}
		}
	}
	
	// Fallback to a generic prompt if role file not found
	fallback := fmt.Sprintf("You are a %s assistant. Help the user with tasks related to your specialization.", roleName)
	fmt.Printf("Using fallback prompt for %s: %s\n", roleName, fallback)
	return fallback, nil
}

func main() {
	prompt, err := loadRolePrompt("coder")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Final prompt: %s\n", prompt)
	}
}
