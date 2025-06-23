package core

import "strings"

func applyVars(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

func applyVarsMap(m map[string]any, vars map[string]string) {
	for k, v := range m {
		switch t := v.(type) {
		case string:
			m[k] = applyVars(t, vars)
		case map[string]any:
			applyVarsMap(t, vars)
		case []any:
			for i, elem := range t {
				switch e := elem.(type) {
				case string:
					t[i] = applyVars(e, vars)
				case map[string]any:
					applyVarsMap(e, vars)
				}
			}
		}
	}
}
