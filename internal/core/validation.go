package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONValidator provides validation for JSON output from agents
type JSONValidator struct {
	// MaxSize limits JSON output size to prevent memory issues
	MaxSize int
	// AllowedRootTypes restricts what can be at JSON root level
	AllowedRootTypes []string
	// DisallowedKeys prevents certain keys that could cause issues
	DisallowedKeys []string
}

// NewJSONValidator creates a validator with sensible defaults
func NewJSONValidator() *JSONValidator {
	return &JSONValidator{
		MaxSize:          10 * 1024 * 1024, // 10MB limit
		AllowedRootTypes: []string{"object", "array"},
		DisallowedKeys:   []string{"__proto__", "constructor", "prototype"},
	}
}

// ValidateToolArgs validates JSON arguments passed to tools
func (v *JSONValidator) ValidateToolArgs(args map[string]any) error {
	// Check for disallowed keys
	for key := range args {
		for _, disallowed := range v.DisallowedKeys {
			if strings.EqualFold(key, disallowed) {
				return fmt.Errorf("disallowed key in tool args: %s", key)
			}
		}
	}

	// Check serialization size
	b, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("tool args not valid JSON: %w", err)
	}
	if len(b) > v.MaxSize {
		return fmt.Errorf("tool args too large: %d bytes (max %d)", len(b), v.MaxSize)
	}

	return nil
}

// ValidateToolResponse validates JSON response from tools
func (v *JSONValidator) ValidateToolResponse(response string) error {
	if len(response) > v.MaxSize {
		return fmt.Errorf("tool response too large: %d bytes (max %d)", len(response), v.MaxSize)
	}

	// Must be valid JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Allow non-JSON responses (many tools return plain text)
		return nil
	}

	// If it is JSON, check root type restrictions
	rootType := getJSONType(parsed)
	if len(v.AllowedRootTypes) > 0 {
		allowed := false
		for _, allowedType := range v.AllowedRootTypes {
			if rootType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("disallowed JSON root type: %s (allowed: %v)", rootType, v.AllowedRootTypes)
		}
	}

	return nil
}

// ValidateAgentOutput validates final output from agents
func (v *JSONValidator) ValidateAgentOutput(output string) error {
	if len(output) > v.MaxSize {
		return fmt.Errorf("agent output too large: %d bytes (max %d)", len(output), v.MaxSize)
	}

	// Check for potential echo patterns (simple heuristic)
	if v.detectEchoPattern(output) {
		return fmt.Errorf("potential echo pattern detected in agent output")
	}

	return nil
}

// detectEchoPattern looks for simple repetitive patterns that might indicate infinite loops
func (v *JSONValidator) detectEchoPattern(output string) bool {
	lines := strings.Split(output, "\n")
	if len(lines) < 10 {
		return false
	}

	// Check if the last 5 lines are identical to previous 5 lines
	lastFive := lines[len(lines)-5:]
	prevFive := lines[len(lines)-10 : len(lines)-5]

	for i := 0; i < 5; i++ {
		if lastFive[i] != prevFive[i] {
			return false
		}
	}

	// If we find identical blocks, it might be an echo
	return true
}

// getJSONType returns the type of a parsed JSON value
func getJSONType(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}
