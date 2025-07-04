package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ToolManifest struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Type        string         `yaml:"type,omitempty" json:"type,omitempty"`
	Command     string         `yaml:"command,omitempty" json:"command,omitempty"`
	HTTP        string         `yaml:"http,omitempty" json:"http,omitempty"`
	Args        map[string]any `yaml:"args,omitempty" json:"args,omitempty"`
}

type ModelManifest struct {
	Name     string            `yaml:"name" json:"name"`
	Provider string            `yaml:"provider" json:"provider"`
	Options  map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

type RouteRule struct {
	IfContains []string `yaml:"if_contains" json:"if_contains"`
	Model      string   `yaml:"model" json:"model"`
}

type File struct {
	Models      []ModelManifest              `yaml:"models" json:"models"`
	Routes      []RouteRule                  `yaml:"routes" json:"routes"`
	Tools       []ToolManifest               `yaml:"tools" json:"tools"`
	Themes      map[string]string            `yaml:"themes" json:"themes"`
	Keybinds    map[string]string            `yaml:"keybinds" json:"keybinds"`
	Credentials map[string]map[string]string `yaml:"credentials" json:"credentials"`
	MCPServers  map[string]string            `yaml:"mcp_servers" json:"mcp_servers"`
}

func merge(dst *File, src File) {
	if len(src.Models) > 0 {
		dst.Models = src.Models
	}
	if len(src.Routes) > 0 {
		dst.Routes = src.Routes
	}
	if len(src.Tools) > 0 {
		dst.Tools = src.Tools
	}
	if dst.Themes == nil {
		dst.Themes = map[string]string{}
	}
	for k, v := range src.Themes {
		dst.Themes[k] = v
	}
	if dst.Keybinds == nil {
		dst.Keybinds = map[string]string{}
	}
	for k, v := range src.Keybinds {
		dst.Keybinds[k] = v
	}
	if dst.Credentials == nil {
		dst.Credentials = map[string]map[string]string{}
	}
	for k, v := range src.Credentials {
		dst.Credentials[k] = v
	}
	if dst.MCPServers == nil {
		dst.MCPServers = map[string]string{}
	}
	for k, v := range src.MCPServers {
		dst.MCPServers[k] = v
	}
}

func Load(path string) (*File, error) {
	var out File

	configHome := os.Getenv("AGENTRY_CONFIG_HOME")
	if configHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configHome = filepath.Join(home, ".config", "agentry")
		}
	}
	if configHome != "" {
		p := filepath.Join(configHome, "config.json")
		if b, err := os.ReadFile(p); err == nil {
			var f File
			if json.Unmarshal(b, &f) == nil {
				merge(&out, f)
			}
		}
	}

	projDir := filepath.Dir(path)
	if projDir == "." || projDir == "" {
		projDir, _ = os.Getwd()
	}
	if b, err := os.ReadFile(filepath.Join(projDir, "agentry.json")); err == nil {
		var f File
		if json.Unmarshal(b, &f) == nil {
			merge(&out, f)
		}
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var yamlFile File
	if err := yaml.Unmarshal(b, &yamlFile); err != nil {
		return nil, err
	}
	merge(&out, yamlFile)
	return &out, nil
}
