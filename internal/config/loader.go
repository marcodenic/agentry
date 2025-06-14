package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ToolManifest struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Command     string         `yaml:"command,omitempty"`
	HTTP        string         `yaml:"http,omitempty"`
	Args        map[string]any `yaml:"args,omitempty"`
}

type File struct {
	OpenAIKey string         `yaml:"openai_key"`
	Tools     []ToolManifest `yaml:"tools"`
}

func Load(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	return &f, yaml.Unmarshal(b, &f)
}
