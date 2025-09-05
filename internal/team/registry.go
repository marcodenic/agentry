package team

import (
	"sort"
	"strings"

	"github.com/marcodenic/agentry/internal/tool"
)

// curatedToolsForRole returns a small, role-specific default tool subset.
func curatedToolsForRole(role string) []string {
	r := strings.ToLower(strings.TrimSpace(role))
	switch r {
	case "coder":
		return []string{
			"read_lines", "view", "edit_range", "create", "search_replace", "insert_at", "fileinfo",
			"bash", "sh",
			"ls", "find", "glob", "grep",
			"patch", "branch-tidy",
			"lsp_diagnostics",
		}
	case "reviewer", "critic", "editor":
		return []string{"view", "read_lines", "lsp_diagnostics"}
	case "tester":
		return []string{"view", "read_lines", "lsp_diagnostics"}
	case "researcher", "writer":
		return []string{"web_search", "read_webpage", "api"}
	default:
		return []string{"view", "read_lines"}
	}
}

// filterRegistryByNames keeps only the specified tools (up to capN) from reg.
func filterRegistryByNames(reg tool.Registry, names []string, capN int) tool.Registry {
	out := make(tool.Registry)
	count := 0
	for _, n := range names {
		if tl, ok := reg[n]; ok {
			out[n] = tl
			count++
			if capN > 0 && count >= capN {
				break
			}
		}
	}
	return out
}

// capRegistry returns at most capN tools from reg in deterministic order.
func capRegistry(reg tool.Registry, capN int) tool.Registry {
	if capN <= 0 {
		return reg
	}
	names := make([]string, 0, len(reg))
	for n := range reg {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make(tool.Registry)
	for i := 0; i < len(names) && i < capN; i++ {
		n := names[i]
		out[n] = reg[n]
	}
	return out
}
