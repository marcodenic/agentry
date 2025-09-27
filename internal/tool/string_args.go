package tool

import "strings"

// stringArg returns the first non-empty string value for the given keys.
func stringArg(args map[string]any, keys ...string) string {
	for _, key := range keys {
		if raw, ok := args[key]; ok {
			if s, ok := raw.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}
