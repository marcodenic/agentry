package memstore

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SharedStore is a simple namespaced key-value store with TTL support.
type SharedStore interface {
	Set(ns, key string, val []byte, ttl time.Duration) error
	Get(ns, key string) ([]byte, bool, error)
	Delete(ns, key string) error
	Keys(ns string) ([]string, error)
	CleanupExpired() error
}

var (
	once         sync.Once
	defaultStore SharedStore
)

// Init initializes the default store once.
func Init() {
	once.Do(func() {
		// Determine backing store from env
		switch os.Getenv("AGENTRY_STORE") {
		case "file":
			root := os.Getenv("AGENTRY_STORE_PATH")
			if root == "" {
				if home, err := os.UserHomeDir(); err == nil {
					root = filepath.Join(home, ".local", "share", "agentry", "store")
				} else {
					root = ".agentry_store"
				}
			}
			defaultStore = NewFileStore(root)
		default:
			defaultStore = NewMemoryStore()
		}
	})
}

// Get returns the initialized default SharedStore.
func Get() SharedStore {
	if defaultStore == nil {
		Init()
	}
	return defaultStore
}
