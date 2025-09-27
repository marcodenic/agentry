package tool

// builtinRegistry collects builtin specs by name without relying on init-time side effects.
type builtinRegistry struct {
	specs map[string]builtinSpec
}

func newBuiltinRegistry() *builtinRegistry {
	return &builtinRegistry{specs: make(map[string]builtinSpec)}
}

func (r *builtinRegistry) addAll(src map[string]builtinSpec) {
	for name, spec := range src {
		r.specs[name] = spec
	}
}

func (r *builtinRegistry) finalize() map[string]builtinSpec {
	return r.specs
}
