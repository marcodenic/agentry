package env

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestWarnDeprecatedEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected []string
	}{
		{
			name:     "no AGENTRY vars",
			envVars:  map[string]string{"OTHER_VAR": "value"},
			expected: nil,
		},
		{
			name:     "only AGENTRY_CONFIG",
			envVars:  map[string]string{"AGENTRY_CONFIG": "/path/to/config.yaml"},
			expected: nil,
		},
		{
			name:     "single deprecated var",
			envVars:  map[string]string{"AGENTRY_FOO": "value"},
			expected: []string{"AGENTRY_FOO"},
		},
		{
			name: "multiple deprecated vars",
			envVars: map[string]string{
				"AGENTRY_FOO": "value1",
				"AGENTRY_BAR": "value2",
			},
			expected: []string{"AGENTRY_BAR", "AGENTRY_FOO"}, // sorted
		},
		{
			name: "mixed valid and deprecated",
			envVars: map[string]string{
				"AGENTRY_CONFIG": "/path/to/config.yaml",
				"AGENTRY_FOO":    "value1",
				"AGENTRY_BAR":    "value2",
				"OTHER_VAR":      "value3",
			},
			expected: []string{"AGENTRY_BAR", "AGENTRY_FOO"}, // sorted, no AGENTRY_CONFIG
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all AGENTRY_* env vars first
			for _, env := range os.Environ() {
				if strings.HasPrefix(env, "AGENTRY_") {
					key := strings.SplitN(env, "=", 2)[0]
					os.Unsetenv(key)
				}
			}

			// Set test env vars
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Call function
			warned := WarnDeprecatedEnvVars()

			// Restore stderr and read output
			w.Close()
			os.Stderr = oldStderr
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check returned slice
			if len(warned) != len(tt.expected) {
				t.Errorf("expected %d warned vars, got %d: %v", len(tt.expected), len(warned), warned)
			}
			for i, expected := range tt.expected {
				if i >= len(warned) || warned[i] != expected {
					t.Errorf("expected warned[%d] = %s, got %v", i, expected, warned)
				}
			}

			// Check stderr output
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(tt.expected) == 0 {
				if output != "" {
					t.Errorf("expected no output, got: %s", output)
				}
			} else {
				if len(lines) != len(tt.expected) {
					t.Errorf("expected %d output lines, got %d: %v", len(tt.expected), len(lines), lines)
				}
				for i, expectedVar := range tt.expected {
					if i >= len(lines) {
						t.Errorf("missing output line for %s", expectedVar)
						continue
					}
					expectedLine := "Deprecation: environment variable " + expectedVar + " is deprecated and ignored. Use YAML config; AGENTRY_CONFIG is the only supported env var."
					if lines[i] != expectedLine {
						t.Errorf("expected line: %s\ngot: %s", expectedLine, lines[i])
					}
				}
			}
		})
	}
}

func TestWarnDeprecatedEnvVars_EdgeCases(t *testing.T) {
	t.Run("AGENTRY_ prefix but empty suffix", func(t *testing.T) {
		// Clear all AGENTRY_* env vars first
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "AGENTRY_") {
				key := strings.SplitN(env, "=", 2)[0]
				os.Unsetenv(key)
			}
		}

		// This shouldn't be possible in practice, but test edge case
		t.Setenv("AGENTRY_", "value")

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		warned := WarnDeprecatedEnvVars()

		w.Close()
		os.Stderr = oldStderr
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Should warn about AGENTRY_ (empty suffix)
		if len(warned) != 1 || warned[0] != "AGENTRY_" {
			t.Errorf("expected to warn about AGENTRY_, got: %v", warned)
		}
		if !strings.Contains(output, "AGENTRY_") {
			t.Errorf("expected output to contain AGENTRY_, got: %s", output)
		}
	})
}
