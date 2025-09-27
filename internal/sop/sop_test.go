package sop

import (
	"strings"
	"testing"
)

func TestGetApplicableSOPsFiltersByRoleAndConditions(t *testing.T) {
	reg := NewSOPRegistry()
	reg.AddSOP(SOP{ID: "match", Roles: []string{"dev"}, Conditions: []string{"env=prod"}})
	reg.AddSOP(SOP{ID: "miss-role", Roles: []string{"ops"}, Conditions: []string{"env=prod"}})
	reg.AddSOP(SOP{ID: "wildcard", Roles: []string{"*"}, Conditions: []string{"feature contains launch"}})

	ctx := map[string]string{"env": "prod", "feature": "launch-ready"}
	applicable := reg.GetApplicableSOPs("dev", ctx)

	if len(applicable) != 2 {
		t.Fatalf("expected 2 applicable SOPs, got %d", len(applicable))
	}

	ids := map[string]bool{}
	for _, sop := range applicable {
		ids[sop.ID] = true
	}
	if !ids["match"] || !ids["wildcard"] {
		t.Fatalf("unexpected SOP selection: %v", ids)
	}
}

func TestEvaluateConditionVariants(t *testing.T) {
	reg := NewSOPRegistry()
	ctx := map[string]string{
		"env":   "prod",
		"notes": "release candidate",
		"flag":  "yes",
	}

	tests := map[string]bool{
		"env=prod":               true,
		"env=dev":                false,
		"notes contains release": true,
		"notes contains beta":    false,
		"flag":                   true,
		"missing":                false,
	}

	for cond, want := range tests {
		if got := reg.evaluateCondition(cond, ctx); got != want {
			t.Fatalf("condition %q: got %v want %v", cond, got, want)
		}
	}
}

func TestFormatSOPsAsPrompt(t *testing.T) {
	reg := NewSOPRegistry()
	reg.AddSOP(SOP{
		ID:          "guidance",
		Title:       "Follow Guidance",
		Description: "Do the right thing",
		Conditions:  []string{"role=coder"},
		Actions:     []string{"Do this", "Then that"},
		Roles:       []string{"coder"},
	})

	ctx := map[string]string{"role": "coder"}
	prompt := reg.FormatSOPsAsPrompt("coder", ctx)

	if !strings.Contains(prompt, "## Standard Operating Procedures") {
		t.Fatalf("prompt missing header: %q", prompt)
	}
	if !strings.Contains(prompt, "### Follow Guidance") {
		t.Fatalf("prompt missing title: %q", prompt)
	}
	if !strings.Contains(prompt, "- Do this") || !strings.Contains(prompt, "- Then that") {
		t.Fatalf("prompt missing actions: %q", prompt)
	}
}

func TestLoadDefaultSOPsAddsExpectedEntries(t *testing.T) {
	reg := NewSOPRegistry()
	reg.LoadDefaultSOPs()

	ctx := map[string]string{"error_occurred": "1"}
	applicable := reg.GetApplicableSOPs("any", ctx)

	found := false
	for _, sop := range applicable {
		if sop.ID == "error-handling" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected default error-handling SOP to be applicable, got %v", applicable)
	}
}
