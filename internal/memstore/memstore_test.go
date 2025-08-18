package memstore

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMemoryStore_SetGetKeysTTL(t *testing.T) {
	ms := NewMemoryStore()
	if err := ms.Set("ns", "a", []byte("1"), 0); err != nil {
		t.Fatal(err)
	}
	if err := ms.Set("ns", "b", []byte("2"), 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	v, ok, err := ms.Get("ns", "a")
	if err != nil || !ok || string(v) != "1" {
		t.Fatalf("get a failed: %v %v %s", err, ok, v)
	}
	keys, err := ms.Keys("ns")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	time.Sleep(20 * time.Millisecond)
	_ = ms.CleanupExpired()
	if _, ok, _ = ms.Get("ns", "b"); ok {
		t.Fatal("expected b expired")
	}
}

func TestFileStore_SetGetKeys(t *testing.T) {
	dir := t.TempDir()
	fs := NewFileStore(dir)
	if err := fs.Set("team1", "k1", []byte("v1"), 0); err != nil {
		t.Fatal(err)
	}
	if err := fs.Set("team1", "k2", []byte("v2"), 0); err != nil {
		t.Fatal(err)
	}
	v, ok, err := fs.Get("team1", "k1")
	if err != nil || !ok || string(v) != "v1" {
		t.Fatalf("get failed: %v %v %s", err, ok, v)
	}
	keys, err := fs.Keys("team1")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	// ensure files exist
	if _, err := os.Stat(filepath.Join(dir, "team1", "k1.json")); err != nil {
		t.Fatal(err)
	}
}
