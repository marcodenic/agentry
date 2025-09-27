package team

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"
)

type stubSharedStore struct {
	mu   sync.Mutex
	data map[string]map[string][]byte
}

func newStubSharedStore() *stubSharedStore {
	return &stubSharedStore{data: make(map[string]map[string][]byte)}
}

func (s *stubSharedStore) Set(ns, key string, val []byte, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[ns]; !ok {
		s.data[ns] = make(map[string][]byte)
	}
	s.data[ns][key] = append([]byte(nil), val...)
	return nil
}

func (s *stubSharedStore) Get(ns, key string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mp, ok := s.data[ns]
	if !ok {
		return nil, false, nil
	}
	val, ok := mp[key]
	if !ok {
		return nil, false, nil
	}
	return append([]byte(nil), val...), true, nil
}

func (s *stubSharedStore) Delete(ns, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if mp, ok := s.data[ns]; ok {
		delete(mp, key)
		if len(mp) == 0 {
			delete(s.data, ns)
		}
	}
	return nil
}

func (s *stubSharedStore) Keys(ns string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	mp, ok := s.data[ns]
	if !ok {
		return nil, nil
	}
	keys := make([]string, 0, len(mp))
	for k := range mp {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *stubSharedStore) CleanupExpired() error { return nil }

func TestSetSharedDataPersistsAndLogs(t *testing.T) {
	store := newStubSharedStore()
	tm := &Team{
		name:         "squad",
		sharedMemory: make(map[string]interface{}),
		store:        store,
		coordination: make([]CoordinationEvent, 0),
	}

	payload := map[string]any{"state": "ready"}
	tm.SetSharedData("status", payload)

	if len(tm.coordination) != 1 {
		t.Fatalf("expected coordination log entry, got %d", len(tm.coordination))
	}
	if tm.coordination[0].Type != "shared_memory_update" {
		t.Fatalf("unexpected coordination event type: %s", tm.coordination[0].Type)
	}
	if stored := tm.sharedMemory["status"]; stored == nil {
		t.Fatalf("expected shared memory to cache value")
	}
	if _, ok := store.data["squad"]["status"]; !ok {
		t.Fatalf("expected backing store to receive value")
	}
}

func TestGetSharedDataCachesFromBackingStore(t *testing.T) {
	store := newStubSharedStore()
	tm := &Team{
		name:         "guild",
		sharedMemory: make(map[string]interface{}),
		store:        store,
	}
	slice := []map[string]any{{"task": "review"}}
	b, err := json.Marshal(slice)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := store.Set("guild", "agenda", b, 0); err != nil {
		t.Fatalf("store set: %v", err)
	}

	val, ok := tm.GetSharedData("agenda")
	if !ok {
		t.Fatalf("expected to load agenda from backing store")
	}
	typed, ok := val.([]map[string]interface{})
	if !ok || len(typed) != 1 {
		t.Fatalf("expected normalized slice of maps, got %#v", val)
	}
	if tm.sharedMemory["agenda"] == nil {
		t.Fatalf("expected value to be cached in shared memory")
	}
}

func TestGetSharedDataReturnsExistingValue(t *testing.T) {
	tm := &Team{
		sharedMemory: map[string]interface{}{"notes": "cached"},
	}

	val, ok := tm.GetSharedData("notes")
	if !ok || val.(string) != "cached" {
		t.Fatalf("expected cached value, got %#v", val)
	}
}

func TestGetAllSharedDataReturnsCopy(t *testing.T) {
	original := map[string]interface{}{"numbers": []int{1, 2, 3}}
	tm := &Team{sharedMemory: original}

	snapshot := tm.GetAllSharedData()
	if !reflect.DeepEqual(snapshot, original) {
		t.Fatalf("expected snapshot to equal original, got %#v", snapshot)
	}

	snapshot["numbers"] = []int{9}
	if reflect.DeepEqual(snapshot, original) {
		t.Fatalf("mutating snapshot should not affect original")
	}
}

func TestNormalizeSharedKeepsMixedSlices(t *testing.T) {
	mixed := []interface{}{map[string]interface{}{"a": 1}, "string"}
	if out := normalizeShared(mixed); !reflect.DeepEqual(out, mixed) {
		t.Fatalf("expected mixed slice to remain unchanged")
	}

	homogenous := []interface{}{map[string]interface{}{"a": 1}}
	converted := normalizeShared(homogenous)
	if _, ok := converted.([]map[string]interface{}); !ok {
		t.Fatalf("expected homogenous slice to convert, got %#v", converted)
	}
}
