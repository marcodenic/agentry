package memstore

import (
	"context"
	"sync"
	"time"
)

// InMemory is a simple ephemeral store backed by Go maps.
type InMemory struct {
	mu   sync.RWMutex
	kv   map[string]map[string][]byte
	meta map[string]map[string]int64
}

func NewInMemory() *InMemory {
	return &InMemory{
		kv:   map[string]map[string][]byte{},
		meta: map[string]map[string]int64{},
	}
}

func (m *InMemory) Set(_ context.Context, bucket, key string, val []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.kv[bucket] == nil {
		m.kv[bucket] = map[string][]byte{}
	}
	if m.meta[bucket] == nil {
		m.meta[bucket] = map[string]int64{}
	}
	cp := make([]byte, len(val))
	copy(cp, val)
	m.kv[bucket][key] = cp
	m.meta[bucket][key] = time.Now().Unix()
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

func (m *InMemory) Cleanup(_ context.Context, bucket string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.kv[bucket] == nil {
		return nil
	}
	before := time.Now().Add(-ttl).Unix()
	for k, ts := range m.meta[bucket] {
		if ts < before {
			delete(m.kv[bucket], k)
			delete(m.meta[bucket], k)
		}
	}
	return nil
}
