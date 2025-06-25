package core

import (
	"gopkg.in/yaml.v3"

	"github.com/marcodenic/agentry/internal/prompts"
)

var defaultSystemPrompt string

func init() {
	var data struct {
		Prompt string `yaml:"prompt"`
	}
	if err := yaml.Unmarshal(prompts.Agent0, &data); err == nil && data.Prompt != "" {
		defaultSystemPrompt = data.Prompt
	} else {
		defaultSystemPrompt = "You are an agent. Use the tools provided to answer the user's question. When you call a tool, `arguments` must be a valid JSON object (use {} if no parameters). Control characters are forbidden."
	}
}

func defaultPrompt() string { return defaultSystemPrompt }
