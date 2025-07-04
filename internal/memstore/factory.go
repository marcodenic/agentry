package memstore

import (
	"fmt"
	"strings"
)

// StoreFactory parses a memory URI and returns the appropriate KV store.
// Supported schemes:
//
//	sqlite:/path/to/db.sqlite
//	file:/path/to/data.json
//	mem:
func StoreFactory(uri string) (KV, error) {
	if uri == "" {
		return nil, nil
	}
	switch {
	case strings.HasPrefix(uri, "sqlite:"):
		path := strings.TrimPrefix(uri, "sqlite:")
		return NewSQLite(path)
	case strings.HasPrefix(uri, "file:"):
		path := strings.TrimPrefix(uri, "file:")
		return NewFile(path)
	case strings.HasPrefix(uri, "mem:") || uri == "mem":
		return NewInMemory(), nil
	default:
		return nil, fmt.Errorf("unknown memory scheme: %s", uri)
	}
}
