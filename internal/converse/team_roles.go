package converse

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RoleConfig represents a role configuration from YAML
type RoleConfig struct {
	Name        string   `yaml:"name"`
	Prompt      string   `yaml:"prompt"`
	Tools       []string `yaml:"tools,omitempty"`       // Legacy support
	Commands    []string `yaml:"commands,omitempty"`    // New semantic commands
	Builtins    []string `yaml:"builtins,omitempty"`    // Allowed builtin tools
	Personality string   `yaml:"personality,omitempty"` // For template substitution
}

// loadRoleConfig loads a complete role configuration from the templates/roles directory
func loadRoleConfig(roleName string) (*RoleConfig, error) {
	if v, ok := roleConfigCache.Load(roleName); ok {
		return v.(*RoleConfig), nil
	}

	roleDirOnce.Do(func() { roleTemplatesDir = findRoleTemplatesDir() })

	if roleTemplatesDir != "" {
		roleFile := filepath.Join(roleTemplatesDir, roleName+".yaml")
		if data, err := os.ReadFile(roleFile); err == nil {
			var config RoleConfig
			if err := yaml.Unmarshal(data, &config); err != nil {
				return nil, err
			}
			roleConfigCache.Store(roleName, &config)
			return &config, nil
		}
	}

	cfg := &RoleConfig{
		Name:   roleName,
		Prompt: fmt.Sprintf("You are a %s assistant. Help the user with tasks related to your specialization.", roleName),
		Tools:  []string{},
	}
	roleConfigCache.Store(roleName, cfg)
	return cfg, nil
}

// loadRolePrompt loads a role-specific prompt from the templates/roles directory (deprecated, use loadRoleConfig)
func loadRolePrompt(roleName string) (string, error) {
	config, err := loadRoleConfig(roleName)
	if err != nil {
		return "", err
	}
	return config.Prompt, nil
}
