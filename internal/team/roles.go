package team

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadRoleFromFile loads a role configuration from a YAML file
func LoadRoleFromFile(path string) (*RoleConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read role file %s: %w", path, err)
	}

	var role RoleConfig
	if err := yaml.Unmarshal(b, &role); err != nil {
		return nil, fmt.Errorf("failed to parse role YAML from %s: %w", path, err)
	}

	return &role, nil
}

// LoadRolesFromIncludePaths loads roles from the include paths in the config
func LoadRolesFromIncludePaths(includePaths []string, configDir string) (map[string]*RoleConfig, error) {
	roles := make(map[string]*RoleConfig)

	for _, includePath := range includePaths {
		// Handle relative paths relative to config directory
		fullPath := includePath
		if !filepath.IsAbs(includePath) {
			fullPath = filepath.Join(configDir, includePath)
		}

		role, err := LoadRoleFromFile(fullPath)
		if err != nil {
			fmt.Printf("Warning: failed to load role from %s: %v\n", fullPath, err)
			continue
		}

		if role.Name == "" {
			fmt.Printf("Warning: role in %s has no name, skipping\n", fullPath)
			continue
		}

		roles[role.Name] = role
		fmt.Printf("✅ Loaded role: %s (model: %v)\n", role.Name, role.Model)
	}

	return roles, nil
}
