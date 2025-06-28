package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	// Enable debug logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[DEBUG] Debug trace script - checking role configurations")

	// Load agent_0 role configuration
	agent0Config, err := loadRoleConfig("agent_0")
	if err != nil {
		log.Printf("[DEBUG] Failed to load agent_0 config: %v", err)
	} else {
		log.Printf("[DEBUG] Loaded agent_0 config:")
		log.Printf("[DEBUG]   Name: %s", agent0Config.Name)
		log.Printf("[DEBUG]   Prompt length: %d chars", len(agent0Config.Prompt))
		log.Printf("[DEBUG]   Commands: %v", agent0Config.Commands)
		log.Printf("[DEBUG]   Builtins: %v", agent0Config.Builtins)
		log.Printf("[DEBUG]   Tools (legacy): %v", agent0Config.Tools)
		log.Printf("[DEBUG]   Prompt preview: %s", agent0Config.Prompt[:min(300, len(agent0Config.Prompt))])
	}

	// Load coder role configuration
	coderConfig, err := loadRoleConfig("coder")
	if err != nil {
		log.Printf("[DEBUG] Failed to load coder config: %v", err)
	} else {
		log.Printf("[DEBUG] Loaded coder config:")
		log.Printf("[DEBUG]   Name: %s", coderConfig.Name)
		log.Printf("[DEBUG]   Prompt length: %d chars", len(coderConfig.Prompt))
		log.Printf("[DEBUG]   Commands: %v", coderConfig.Commands)
		log.Printf("[DEBUG]   Builtins: %v", coderConfig.Builtins)
		log.Printf("[DEBUG]   Tools (legacy): %v", coderConfig.Tools)
		log.Printf("[DEBUG]   Prompt preview: %s", coderConfig.Prompt[:min(300, len(coderConfig.Prompt))])
	}

	log.Println("\n[DEBUG] Now run agentry.exe in TUI mode to see debug output during agent creation!")
	log.Println("[DEBUG] Try: .\\agentry.exe tui")
	log.Println("[DEBUG] Then in TUI type: coder read TODO.md")
}

// RoleConfig represents a role configuration from YAML (copied from team.go for this test)
type RoleConfig struct {
	Name        string   `yaml:"name"`
	Prompt      string   `yaml:"prompt"`
	Tools       []string `yaml:"tools,omitempty"`       // Legacy support
	Commands    []string `yaml:"commands,omitempty"`    // New semantic commands
	Builtins    []string `yaml:"builtins,omitempty"`    // Allowed builtin tools
	Personality string   `yaml:"personality,omitempty"` // For template substitution
}

// loadRoleConfig loads a complete role configuration from the templates/roles directory (copied from team.go for this test)
func loadRoleConfig(roleName string) (*RoleConfig, error) {
	// Try to find the templates/roles directory by searching up from the current directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	
	for dir := workDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		templatesDir := filepath.Join(dir, "templates", "roles")
		if _, err := os.Stat(templatesDir); err == nil {
			roleFile := filepath.Join(templatesDir, roleName+".yaml")
			if _, err := os.Stat(roleFile); err == nil {
				data, err := os.ReadFile(roleFile)
				if err != nil {
					return nil, err
				}
				
				var config RoleConfig
				if err := yaml.Unmarshal(data, &config); err != nil {
					return nil, err
				}
				
				return &config, nil
			}
		}
	}
	
	// Fallback to a generic config if role file not found
	return &RoleConfig{
		Name:   roleName,
		Prompt: fmt.Sprintf("You are a %s assistant. Help the user with tasks related to your specialization.", roleName),
		Tools:  []string{}, // Empty tools list for unknown roles
	}, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
