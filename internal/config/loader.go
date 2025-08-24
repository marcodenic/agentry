package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ToolManifest struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Type        string          `yaml:"type,omitempty"`
	Command     string          `yaml:"command,omitempty"`
	HTTP        string          `yaml:"http,omitempty"`
	Args        map[string]any  `yaml:"args,omitempty"`
	Privileged  bool            `yaml:"privileged,omitempty"`
	Net         string          `yaml:"net,omitempty"`
	CPULimit    string          `yaml:"cpu_limit,omitempty"`
	MemLimit    string          `yaml:"mem_limit,omitempty"`
	Engine      string          `yaml:"engine,omitempty"`
	Permissions ToolPermissions `yaml:"permissions,omitempty"`
}

type ToolPermissions struct {
	Allow *bool `yaml:"allow"`
}

type ModelManifest struct {
	Name     string            `yaml:"name"`
	Provider string            `yaml:"provider"`
	Options  map[string]string `yaml:"options,omitempty"`
}

// VectorManifest describes a VectorStore backend.
type VectorManifest struct {
	Type       string `yaml:"type"`
	URL        string `yaml:"url"`
	Collection string `yaml:"collection,omitempty"`
}

type File struct {
	Models      []ModelManifest              `yaml:"models"`
	Tools       []ToolManifest               `yaml:"tools"`
	Include     []string                     `yaml:"include"` // Add include support for role files
	Memory      string                       `yaml:"memory"`
	Store       string                       `yaml:"store"`
	Vector      VectorManifest               `yaml:"vector_store"`
	Theme       string                       `yaml:"theme"`
	Themes      map[string]string            `yaml:"themes"`
	Keybinds    map[string]string            `yaml:"keybinds"`
	Credentials map[string]map[string]string `yaml:"credentials"`
	MCPServers  map[string]string            `yaml:"mcp_servers"`
	Collector   string                       `yaml:"collector"`
	Port        string                       `yaml:"port"`
	Sandbox     Sandbox                      `yaml:"sandbox"`
	Permissions Permissions                  `yaml:"permissions"`
	Budget      Budget                       `yaml:"budget"`
}

type Sandbox struct {
	Engine string `yaml:"engine"`
}

type Permissions struct {
	Tools []string `yaml:"tools"`
}

type Budget struct {
	Tokens  int     `yaml:"tokens"`
	Dollars float64 `yaml:"dollars"`
}

// Validate performs basic sanity checks on the loaded configuration.
func (f *File) Validate() error {
	// No validation needed currently - models and tools are validated separately
	return nil
}

func merge(dst *File, src File) {
	if len(src.Models) > 0 {
		dst.Models = src.Models
	}
	if len(src.Tools) > 0 {
		dst.Tools = src.Tools
	}
	if len(src.Include) > 0 {
		dst.Include = src.Include
	}
	if src.Memory != "" {
		dst.Memory = src.Memory
	}
	if src.Store != "" {
		dst.Store = src.Store
	}
	if src.Vector.Type != "" {
		dst.Vector = src.Vector
	}
	if src.Theme != "" {
		dst.Theme = src.Theme
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
	if src.Collector != "" {
		dst.Collector = src.Collector
	}
	if src.Port != "" {
		dst.Port = src.Port
	}
	if src.Sandbox.Engine != "" {
		dst.Sandbox = src.Sandbox
	}
	if len(src.Permissions.Tools) > 0 {
		dst.Permissions = src.Permissions
	}
	if src.Budget.Tokens > 0 || src.Budget.Dollars > 0 {
		dst.Budget = src.Budget
	}
}

func Load(path string) (*File, error) {
	var out File

	// Load global config from user config directory if it exists
	configHome := os.Getenv("AGENTRY_CONFIG_HOME")
	if configHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configHome = filepath.Join(home, ".config", "agentry")
		}
	}
	if configHome != "" {
		globalConfigPath := filepath.Join(configHome, "config.yaml")
		if b, err := os.ReadFile(globalConfigPath); err == nil {
			var f File
			if yaml.Unmarshal(b, &f) == nil {
				merge(&out, f)
			}
		}
	}

	// Load project-level config if it exists
	projDir := filepath.Dir(path)
	if projDir == "." || projDir == "" {
		projDir, _ = os.Getwd()
	}
	if b, err := os.ReadFile(filepath.Join(projDir, "agentry.yaml")); err == nil {
		var f File
		if yaml.Unmarshal(b, &f) == nil {
			merge(&out, f)
		}
	}

	// Load the main config file
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var yamlFile File
	if err := yaml.Unmarshal(b, &yamlFile); err != nil {
		return nil, err
	}
	merge(&out, yamlFile)

	// Apply environment variable overrides
	if v := os.Getenv("AGENTRY_COLLECTOR"); v != "" {
		out.Collector = v
	}
	if v := os.Getenv("AGENTRY_PORT"); v != "" {
		out.Port = v
	}

	return &out, nil
}
