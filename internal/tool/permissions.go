package tool

import (
	"os"

	"gopkg.in/yaml.v3"
)

// permissionsFile describes a YAML permissions list.
type permissionsFile struct {
	Tools []string `yaml:"tools"`
}

// LoadPermissionsFile reads a YAML file and configures allowed tools.
func LoadPermissionsFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pf permissionsFile
	if err := yaml.Unmarshal(b, &pf); err != nil {
		return err
	}
	SetPermissions(pf.Tools)
	return nil
}
