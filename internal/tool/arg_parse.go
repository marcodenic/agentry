package tool

// getIntArg retrieves an integer argument from a generic args map, accepting
// both JSON-style float64 and native int types. Returns (value, okFound).
func getIntArg(args map[string]any, key string, def int) (int, bool) {
	v, ok := args[key]
	if !ok || v == nil {
		return def, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int8:
		return int(n), true
	case int16:
		return int(n), true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case uint:
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		return int(n), true
	case uint64:
		return int(n), true
	case float32:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return def, false
	}
}
