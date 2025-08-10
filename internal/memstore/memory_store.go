package memstore

import (
	"sync"
	"time"
)

type memEntry struct {
	val []byte
	exp time.Time // zero means no expiry
}

// memoryStore is a simple in-memory implementation of SharedStore.
type memoryStore struct {
	mu   sync.RWMutex
	data map[string]map[string]memEntry // ns -> key -> entry
}

// NewMemoryStore creates an in-memory SharedStore.
func NewMemoryStore() SharedStore {
	return &memoryStore{data: make(map[string]map[string]memEntry)}
}

func (m *memoryStore) Set(ns, key string, val []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[ns]; !ok {
		m.data[ns] = make(map[string]memEntry)
	}
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.data[ns][key] = memEntry{val: append([]byte(nil), val...), exp: exp}
	return nil
}

func (m *memoryStore) Get(ns, key string) ([]byte, bool, error) {
	m.mu.RLock()
	entry, ok := m.data[ns][key]
	m.mu.RUnlock()
	if !ok {
		return nil, false, nil
	}
	if !entry.exp.IsZero() && time.Now().After(entry.exp) {
		// expired; cleanup lazily
		m.mu.Lock()
		delete(m.data[ns], key)
		m.mu.Unlock()
		return nil, false, nil
	}
	return append([]byte(nil), entry.val...), true, nil
}

func (m *memoryStore) Delete(ns, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[ns]; ok {
		delete(m.data[ns], key)
		if len(m.data[ns]) == 0 {
			delete(m.data, ns)
		}
	}
	return nil
}

func (m *memoryStore) Keys(ns string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]string, 0)
	if mp, ok := m.data[ns]; ok {
		now := time.Now()
		for k, e := range mp {
			if e.exp.IsZero() || now.Before(e.exp) {
				res = append(res, k)
			}
		}
	}
	return res, nil
}

func (m *memoryStore) CleanupExpired() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for ns, mp := range m.data {
		for k, e := range mp {
			if !e.exp.IsZero() && now.After(e.exp) {
				delete(mp, k)
			}
		}
		if len(mp) == 0 {
			delete(m.data, ns)
		}
	}
	return nil
}
