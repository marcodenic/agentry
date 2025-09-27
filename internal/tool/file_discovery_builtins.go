package tool

func getFileDiscoveryBuiltins() map[string]builtinSpec {
	return fileDiscoveryTools()
}

func registerFileDiscoveryBuiltins(reg *builtinRegistry) {
	reg.addAll(getFileDiscoveryBuiltins())
}
