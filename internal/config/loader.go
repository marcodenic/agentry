package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ToolManifest struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Type        string         `yaml:"type,omitempty"`
	Command     string         `yaml:"command,omitempty"`
	HTTP        string         `yaml:"http,omitempty"`
	Args        map[string]any `yaml:"args,omitempty"`
}

type ModelManifest struct {
	Name     string            `yaml:"name"`
	Provider string            `yaml:"provider"`
	Options  map[string]string `yaml:"options,omitempty"`
}

type RouteRule struct {
	IfContains []string `yaml:"if_contains"`
	Model      string   `yaml:"model"`
}

type File struct {
	Models []ModelManifest `yaml:"models"`
	Routes []RouteRule     `yaml:"routes"`
	Tools  []ToolManifest  `yaml:"tools"`
}

func Load(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
