package tool

func teamTools() map[string]builtinSpec {
	specs := map[string]builtinSpec{
		"team": teamPlanSpec(),
	}
	mergeBuiltinSpecs(specs, teamStatusBuiltins())
	mergeBuiltinSpecs(specs, sharedMemoryBuiltins())
	mergeBuiltinSpecs(specs, coordinationBuiltins())
	return specs
}

func mergeBuiltinSpecs(dst, src map[string]builtinSpec) {
	for name, spec := range src {
		dst[name] = spec
	}
}
