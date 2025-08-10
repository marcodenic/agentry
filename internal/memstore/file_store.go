package memstore

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileStore struct {
	root string
	mu   sync.RWMutex
}

type fileRecord struct {
	Value     []byte    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewFileStore creates a directory-backed SharedStore.
func NewFileStore(root string) SharedStore {
	_ = os.MkdirAll(root, 0o755)
	return &fileStore{root: root}
}

func (f *fileStore) nsDir(ns string) string {
	return filepath.Join(f.root, ns)
}

func (f *fileStore) filePath(ns, key string) string {
	// keep key as filename (safe subset expected). If contains path sep, replace.
	sanitized := filepath.Base(key)
	return filepath.Join(f.nsDir(ns), sanitized+".json")
}

func (f *fileStore) Set(ns, key string, val []byte, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if ns == "" || key == "" {
		return errors.New("namespace and key required")
	}
	if err := os.MkdirAll(f.nsDir(ns), 0o755); err != nil {
		return err
	}
	rec := fileRecord{}
	// store raw bytes but wrap in json for metadata
	rec.Value = val
	if ttl > 0 {
		rec.ExpiresAt = time.Now().Add(ttl)
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(f.filePath(ns, key), b, 0o644)
}

func (f *fileStore) Get(ns, key string) ([]byte, bool, error) {
	f.mu.RLock()
	path := f.filePath(ns, key)
	f.mu.RUnlock()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var rec fileRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return nil, false, err
	}
	if !rec.ExpiresAt.IsZero() && time.Now().After(rec.ExpiresAt) {
		_ = os.Remove(path)
		return nil, false, nil
	}
	return rec.Value, true, nil
}

func (f *fileStore) Delete(ns, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	path := f.filePath(ns, key)
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (f *fileStore) Keys(ns string) ([]string, error) {
	dir := f.nsDir(ns)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	res := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) == ".json" {
			res = append(res, name[:len(name)-5])
		}
	}
	return res, nil
}

func (f *fileStore) CleanupExpired() error {
	// simple scan
	namespaces, err := os.ReadDir(f.root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	now := time.Now()
	for _, ns := range namespaces {
		if !ns.IsDir() {
			continue
		}
		dir := filepath.Join(f.root, ns.Name())
		files, _ := os.ReadDir(dir)
		for _, file := range files {
			if filepath.Ext(file.Name()) != ".json" {
				continue
			}
			path := filepath.Join(dir, file.Name())
			b, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var rec fileRecord
			if err := json.Unmarshal(b, &rec); err != nil {
				_ = os.Remove(path)
				continue
			}
			if !rec.ExpiresAt.IsZero() && now.After(rec.ExpiresAt) {
				_ = os.Remove(path)
			}
		}
	}
	return nil
}
