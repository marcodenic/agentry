// Package tool - builtins.go
//
// This file has been refactored for maintainability. The original builtin tool
// definitions have been split across multiple focused files:
//
// - builtins_utility.go : Basic utility tools (echo, ping)
// - builtins_network.go : Network-related tools (fetch, mcp)
// - builtins_team.go    : Team coordination tools (agent, team_status, etc.)
// - builtins_system.go  : System information tools (sysinfo, project_tree)
//
// This improves code organization and makes the builtin tools easier to maintain.

package tool

// builtinSpec defines builtin schema and execution.
type builtinSpec struct {
	Desc   string
	Schema map[string]any
	Exec   ExecFn
}

// builtinMap holds safe builtin tools keyed by name.
var builtinMap = map[string]builtinSpec{}

// init initializes the builtin map by merging all builtin tool categories
func init() {
	// Merge all builtin tool categories
	for name, spec := range getUtilityBuiltins() {
		builtinMap[name] = spec
	}
	for name, spec := range getNetworkBuiltins() {
		builtinMap[name] = spec
	}
	for name, spec := range getTeamBuiltins() {
		builtinMap[name] = spec
	}
	for name, spec := range getSystemBuiltins() {
		builtinMap[name] = spec
	}
}
