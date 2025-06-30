package registry

import (
	"encoding/json"
	"fmt"
	"os"
)

// loadFromFile loads agents from the JSON file
func (r *FileRegistry) loadFromFile() error {
	data, err := os.ReadFile(r.configFile)
	if err != nil {
		return err
	}

	var fileData struct {
		Agents map[string]*AgentInfo    `json:"agents"`
		Health map[string]*HealthMetrics `json:"health"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}

	if fileData.Agents != nil {
		r.agents = fileData.Agents
	}
	if fileData.Health != nil {
		r.health = fileData.Health
	}

	return nil
}

// saveToFile saves agents to the JSON file
func (r *FileRegistry) saveToFile() error {
	fileData := struct {
		Agents map[string]*AgentInfo    `json:"agents"`
		Health map[string]*HealthMetrics `json:"health"`
	}{
		Agents: r.agents,
		Health: r.health,
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}

	// Write to temporary file first, then rename for atomic update
	tempFile := r.configFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, r.configFile); err != nil {
		os.Remove(tempFile) // Clean up on failure
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
