package memstore

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"
)

// File is a simple JSON-backed store persisted to disk.
type File struct {
	path   string
	mu     sync.RWMutex
	kv     map[string]map[string][]byte
	meta   map[string]map[string]int64
	vector map[string]string
}

func NewFile(path string) (*File, error) {
	f := &File{
		path:   path,
		kv:     map[string]map[string][]byte{},
		meta:   map[string]map[string]int64{},
		vector: map[string]string{},
	}
	if b, err := os.ReadFile(path); err == nil {
		var data struct {
			KV     map[string]map[string][]byte `json:"kv"`
			Meta   map[string]map[string]int64  `json:"meta"`
			Vector map[string]string            `json:"vector"`
		}
		if err := json.Unmarshal(b, &data); err == nil && data.KV != nil {
			f.kv = data.KV
			if data.Meta != nil {
				f.meta = data.Meta
			}
			if data.Vector != nil {
				f.vector = data.Vector
			}
		} else {
			// legacy format: just kv map
			_ = json.Unmarshal(b, &f.kv)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return f, nil
}

func (f *File) persist() error {
	data := struct {
		KV     map[string]map[string][]byte `json:"kv"`
		Meta   map[string]map[string]int64  `json:"meta"`
		Vector map[string]string            `json:"vector"`
	}{f.kv, f.meta, f.vector}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(f.path, b, 0644)
}

func (f *File) Set(_ context.Context, bucket, key string, val []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.kv[bucket] == nil {
		f.kv[bucket] = map[string][]byte{}
	}
	if f.meta[bucket] == nil {
		f.meta[bucket] = map[string]int64{}
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	f.kv[bucket][key] = cp
	f.meta[bucket][key] = time.Now().Unix()
	return f.persist()
}

func (f *File) Get(_ context.Context, bucket, key string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.kv[bucket] == nil {
		return nil, nil
	}
	val, ok := f.kv[bucket][key]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	return cp, nil
}

func (f *File) Add(_ context.Context, id, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.vector[id] = text
	return f.persist()
}

func (f *File) Query(_ context.Context, _ string, k int) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	ids := make([]string, 0, k)
	for id := range f.vector {
		ids = append(ids, id)
		if len(ids) >= k {
			break
		}
	}
	return ids, nil
}

func (f *File) Cleanup(_ context.Context, bucket string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.kv[bucket] == nil {
		return nil
	}
	before := time.Now().Add(-ttl).Unix()
	for k, ts := range f.meta[bucket] {
		if ts < before {
			delete(f.kv[bucket], k)
			delete(f.meta[bucket], k)
		}
	}
	return f.persist()
}
