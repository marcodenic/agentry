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

// builtinModule registers builtin specs with the provided registry.
type builtinModule func(*builtinRegistry)

// builtinModules enumerates all builtin providers in deterministic assembly order.
var builtinModules = []builtinModule{
	registerUtilityBuiltins,
	registerNetworkBuiltins,
	registerWebSearchBuiltins,
	registerTeamBuiltins,
	registerSystemBuiltins,
	registerFileDiscoveryBuiltins,
	registerTodoBuiltins,
	registerAPIBuiltins,
	registerCreateBuiltins,
	registerDownloadBuiltins,
	registerEditRangeBuiltins,
	registerFileInfoBuiltins,
	registerInsertAtBuiltins,
	registerLSPBuiltins,
	registerReadLinesBuiltins,
	registerReadWebpageBuiltins,
	registerSearchReplaceBuiltins,
	registerShellBuiltins,
	registerViewBuiltins,
	registerWriteEditBuiltins,
}

// builtinMap holds safe builtin tools keyed by name.
var builtinMap = assembleBuiltinRegistry()

// assembleBuiltinRegistry collects builtin specs from each module.
func assembleBuiltinRegistry() map[string]builtinSpec {
	reg := newBuiltinRegistry()
	for _, module := range builtinModules {
		module(reg)
	}
	return reg.finalize()
}
