package core

import (
	"os"

	"gopkg.in/yaml.v3"
)

// RoleConfig represents a role configuration (duplicated to avoid import cycle)
type roleConfig struct {
	Name   string `yaml:"name"`
	Prompt string `yaml:"prompt"`
}

// GetDefaultPrompt loads the canonical agent_0 prompt from the role file
func GetDefaultPrompt() string {
	b, err := os.ReadFile("templates/roles/agent_0.yaml")
	if err != nil {
		return "You are Agent 0, the system orchestrator. You can delegate to specialized agents using the agent tool."
	}

	var role roleConfig
	if err := yaml.Unmarshal(b, &role); err != nil {
		return "You are Agent 0, the system orchestrator. You can delegate to specialized agents using the agent tool."
	}

	return role.Prompt
}

func defaultPrompt() string {
	return GetDefaultPrompt()
}
