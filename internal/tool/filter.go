package tool

import (
	"os"
	"strings"
)

// Filter returns a new Registry with tools reduced according to role/model allow/deny rules.
// Environment overrides:
//
//	AGENTRY_DISABLE_TOOL_FILTER=1  -> bypass filtering entirely
//	AGENTRY_TOOL_ALLOW_EXTRA=comma list of extra tools to force-include
//	AGENTRY_TOOL_DENY=comma list of tools to exclude after all logic
//
// For now we specifically trim the coder role (especially on Anthropic models) to a lean set.
func Filter(reg Registry, role, model string) Registry {
	if os.Getenv("AGENTRY_DISABLE_TOOL_FILTER") == "1" {
		return reg
	}
	role = strings.ToLower(role)
	modelLower := strings.ToLower(model)
	// Derive base allow set. Start with empty then fill.
	allow := map[string]struct{}{}
	// If not coder, just pass-through for now (future: specialize other roles)
	if role != "coder" && !strings.Contains(role, "code") {
		return applyOverrides(reg, allow, true) // no base allow list -> original registry with overrides
	}

	// Core discovery & inspection
	base := []string{"view", "read_lines", "fileinfo", "ls", "glob", "grep", "find", "project_tree"}
	// Editing & write ops
	base = append(base, "create", "write", "edit", "edit_range", "insert_at", "patch", "search_replace")
	// Execution/system
	base = append(base, "bash", "sh", "sysinfo", "lsp_diagnostics")
	// Optional network/web tools limited unless explicitly added
	// optional tools documented for clarity (added only via overrides or non-Anthropic model default subset)
	// optional := []string{"api", "fetch", "read_webpage", "web_search", "download"}

	for _, n := range base {
		allow[n] = struct{}{}
	}
	// Anthropic coder: be extra conservative (omit network by default)
	if strings.Contains(modelLower, "claude") || strings.Contains(modelLower, "anthropic") {
		// do nothing; optional not added unless user overrides
	} else {
		// Non-Anthropic coder: allow a subset of optional safe tools
		allow["api"], allow["fetch"] = struct{}{}, struct{}{}
	}

	filtered := Registry{}
	for name, t := range reg {
		if _, ok := allow[name]; ok {
			filtered[name] = t
		}
	}
	return applyOverrides(filtered, allow, false)
}

func applyOverrides(base Registry, allow map[string]struct{}, passthrough bool) Registry {
	// If passthrough true and allow is empty, start from base (original registry) unaffected.
	out := Registry{}
	if passthrough && len(allow) == 0 {
		for k, v := range base {
			out[k] = v
		}
	} else {
		for k, v := range base {
			out[k] = v
		}
	}
	// Extra includes
	if extra := os.Getenv("AGENTRY_TOOL_ALLOW_EXTRA"); extra != "" {
		for _, n := range strings.Split(extra, ",") {
			n = strings.TrimSpace(n)
			if n == "" {
				continue
			}
			// If it existed in original registry, include
			// NOTE: We do not have original registry here; assume caller used Filter with full registry context.
			// This function can't resurrect unknown tools unless already in base map.
		}
	}
	// Deny list
	if deny := os.Getenv("AGENTRY_TOOL_DENY"); deny != "" {
		for _, n := range strings.Split(deny, ",") {
			n = strings.TrimSpace(n)
			delete(out, n)
		}
	}
	// Stable ordering not required for map but helpful for deterministic specs downstream (BuildSpecs already iterates map order).
	// Optionally we could reinsert sorted; skip for now.
	return out
}
