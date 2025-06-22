package tool

// Register allows plugins to add custom builtin tools via init().
func Register(name, desc string, schema map[string]any, fn ExecFn) {
	builtinMap[name] = builtinSpec{Desc: desc, Schema: schema, Exec: fn}
}
