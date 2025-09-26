package sop

import (
	"fmt"
	"strings"
)

// SOP represents a Standard Operating Procedure
type SOP struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Conditions  []string `yaml:"conditions"`
	Actions     []string `yaml:"actions"`
	Priority    int      `yaml:"priority"` // Higher priority SOPs are checked first
	Roles       []string `yaml:"roles"`    // Which roles this SOP applies to
}

// SOPRegistry manages a collection of SOPs
type SOPRegistry struct {
	sops map[string]SOP
}

// NewSOPRegistry creates a new SOP registry
func NewSOPRegistry() *SOPRegistry {
	return &SOPRegistry{
		sops: make(map[string]SOP),
	}
}

// AddSOP adds a SOP to the registry
func (r *SOPRegistry) AddSOP(sop SOP) {
	r.sops[sop.ID] = sop
}

// GetApplicableSOPs returns SOPs that match the given context
func (r *SOPRegistry) GetApplicableSOPs(role string, context map[string]string) []SOP {
	var applicable []SOP

	for _, sop := range r.sops {
		// Check if SOP applies to this role
		if len(sop.Roles) > 0 {
			roleMatches := false
			for _, sopRole := range sop.Roles {
				if sopRole == role || sopRole == "*" {
					roleMatches = true
					break
				}
			}
			if !roleMatches {
				continue
			}
		}

		// Check if conditions are met
		if r.conditionsMatch(sop.Conditions, context) {
			applicable = append(applicable, sop)
		}
	}

	return applicable
}

// conditionsMatch checks if SOP conditions are satisfied by the context
func (r *SOPRegistry) conditionsMatch(conditions []string, context map[string]string) bool {
	for _, condition := range conditions {
		if !r.evaluateCondition(condition, context) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition against context
func (r *SOPRegistry) evaluateCondition(condition string, context map[string]string) bool {
	// Simple condition evaluation - can be extended for complex rules
	if strings.Contains(condition, "=") {
		parts := strings.SplitN(condition, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			return context[key] == value
		}
	}

	if strings.Contains(condition, "contains") {
		parts := strings.SplitN(condition, " contains ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			return strings.Contains(context[key], value)
		}
	}

	// Default: check if the condition key exists in context
	return context[condition] != ""
}

// FormatSOPsAsPrompt formats applicable SOPs as a prompt addition
func (r *SOPRegistry) FormatSOPsAsPrompt(role string, context map[string]string) string {
	sops := r.GetApplicableSOPs(role, context)
	if len(sops) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Standard Operating Procedures\n\n")
	sb.WriteString("The following procedures apply to your current role and context:\n\n")

	for _, sop := range sops {
		sb.WriteString(fmt.Sprintf("### %s\n", sop.Title))
		if sop.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", sop.Description))
		}

		if len(sop.Actions) > 0 {
			sb.WriteString("**Required actions:**\n")
			for _, action := range sop.Actions {
				sb.WriteString(fmt.Sprintf("- %s\n", action))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// LoadDefaultSOPs loads built-in SOPs for common scenarios
func (r *SOPRegistry) LoadDefaultSOPs() {
	// Error handling SOP
	r.AddSOP(SOP{
		ID:          "error-handling",
		Title:       "Error Handling Procedure",
		Description: "How to handle tool execution errors and failures",
		Conditions:  []string{"error_occurred"},
		Actions: []string{
			"Analyze the error message carefully",
			"Check if the tool arguments are correct",
			"Try alternative approaches before giving up",
			"Provide clear error context to the user",
			"Do not repeat the same failing action",
		},
		Priority: 10,
		Roles:    []string{"*"},
	})

	// JSON output validation SOP
	r.AddSOP(SOP{
		ID:          "json-output",
		Title:       "JSON Output Guidelines",
		Description: "Ensuring proper JSON output formatting",
		Conditions:  []string{"output_format=json"},
		Actions: []string{
			"Always validate JSON syntax before outputting",
			"Use proper escaping for special characters",
			"Ensure JSON objects have required fields",
			"Keep JSON responses under size limits",
			"Never include executable code in JSON strings",
		},
		Priority: 8,
		Roles:    []string{"*"},
	})

	// Echo prevention SOP
	r.AddSOP(SOP{
		ID:          "echo-prevention",
		Title:       "Echo Pattern Prevention",
		Description: "Preventing infinite loops and repetitive outputs",
		Conditions:  []string{"iteration_count>5"},
		Actions: []string{
			"Check if you're repeating the same actions",
			"Verify that tool calls are producing new information",
			"Avoid calling the same tool with identical arguments",
			"If stuck, summarize progress and ask for guidance",
		},
		Priority: 9,
		Roles:    []string{"*"},
	})

	// Testing SOP
	r.AddSOP(SOP{
		ID:          "testing-procedure",
		Title:       "Testing Best Practices",
		Description: "Guidelines for running and analyzing tests",
		Conditions:  []string{"role=tester"},
		Actions: []string{
			"Always run tests before making changes",
			"Read test failures carefully for root cause",
			"Run diagnostics when tests fail",
			"Verify fixes by re-running tests",
			"Document test results clearly",
		},
		Priority: 7,
		Roles:    []string{"tester", "coder"},
	})

	// Code modification SOP
	r.AddSOP(SOP{
		ID:          "code-modification",
		Title:       "Code Modification Guidelines",
		Description: "Safe practices for modifying code",
		Conditions:  []string{"role=coder"},
		Actions: []string{
			"Always read files before modifying them",
			"Make small, focused changes",
			"Preserve existing formatting and style",
			"Add comments for complex logic",
			"Test changes after implementation",
		},
		Priority: 6,
		Roles:    []string{"coder", "editor"},
	})

	// Tool usage SOP
	r.AddSOP(SOP{
		ID:          "tool-usage",
		Title:       "Proper Tool Usage",
		Description: "Guidelines for effective tool utilization",
		Conditions:  []string{},
		Actions: []string{
			"Read tool documentation before first use",
			"Provide all required parameters",
			"Use appropriate tools for each task",
			"Handle tool errors gracefully",
			"Don't use tools unnecessarily",
		},
		Priority: 5,
		Roles:    []string{"*"},
	})
}
