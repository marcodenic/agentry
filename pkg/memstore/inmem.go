package memstore

import (
	"context"
	"sync"
)

// InMemory is a simple ephemeral store backed by Go maps.
type InMemory struct {
	mu     sync.RWMutex
	kv     map[string]map[string][]byte
	vector map[string]string
}

func NewInMemory() *InMemory {
	return &InMemory{
		kv:     map[string]map[string][]byte{},
		vector: map[string]string{},
	}
}

func (m *InMemory) Set(_ context.Context, bucket, key string, val []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.kv[bucket] == nil {
		m.kv[bucket] = map[string][]byte{}
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	m.kv[bucket][key] = cp
	return nil
}

func (m *InMemory) Get(_ context.Context, bucket, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.kv[bucket] == nil {
		return nil, nil
	}
	val, ok := m.kv[bucket][key]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	return cp, nil
}

func (m *InMemory) Add(_ context.Context, id, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.vector[id] = text
	return nil
}

func (m *InMemory) Query(_ context.Context, _ string, k int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, k)
	for id := range m.vector {
		ids = append(ids, id)
		if len(ids) >= k {
			break
		}
	}
	return ids, nil
}
