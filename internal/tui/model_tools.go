package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleToolUseMessage processes tool usage messages (tool completion)
func (m Model) handleToolUseMessage(msg toolUseMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	info.CurrentTool = msg.name

	// Complete the progressive status update (add green tick and change bar color)
	info.completeProgressiveStatusUpdate(m)

	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// handleActionMessage processes action notification messages (tool start)
func (m Model) handleActionMessage(msg actionMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]

	// Start progressive status update with orange bar
	info.startProgressiveStatusUpdate(msg.text, m)

	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// handleModelMessage processes model information messages
func (m Model) handleModelMessage(msg modelMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]

	// Determine if the new model name is more specific than the current one
	if shouldUpdateModelName(info.ModelName, msg.name) {
		info.ModelName = msg.name
	}

	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// shouldUpdateModelName determines if the new model name is more specific than the current one
func shouldUpdateModelName(currentName, newName string) bool {
	// If no current name, accept any new name
	if currentName == "" {
		return true
	}

	// Calculate specificity scores for both names
	currentScore := getModelNameSpecificity(currentName)
	newScore := getModelNameSpecificity(newName)

	// Only update if the new name is more specific
	return newScore > currentScore
}

// getModelNameSpecificity returns a score indicating how specific a model name is
// Higher scores are more specific (actual model names vs rule names)
func getModelNameSpecificity(name string) int {
	if name == "" {
		return 0
	}

	// Common router rule names get low scores
	commonRuleNames := map[string]int{
		"openai":    1,
		"anthropic": 1,
		"mock":      1,
		"ollama":    1,
		"gemini":    1,
	}

	if score, exists := commonRuleNames[name]; exists {
		return score
	}

	// Specific model identifiers get higher scores
	specificity := 2 // Base score for non-rule names

	// Increase score for model names with version numbers
	if strings.Contains(name, "-") {
		specificity += 2 // e.g., "gpt-4", "claude-3"
	}

	// Increase score for model names with detailed versions
	if strings.Contains(name, "mini") || strings.Contains(name, "max") ||
		strings.Contains(name, "ultra") || strings.Contains(name, "plus") {
		specificity += 1 // e.g., "gpt-4o-mini"
	}

	// Increase score for model names with specific version numbers
	for _, char := range name {
		if char >= '0' && char <= '9' {
			specificity += 1
			break // Only count once for containing numbers
		}
	}

	return specificity
}
