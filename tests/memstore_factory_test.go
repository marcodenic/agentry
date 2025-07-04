//go:build integration
// +build integration

package tests

import (
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/memstore"
)

func TestStoreFactory(t *testing.T) {
	tmp := t.TempDir()

	s, err := memstore.StoreFactory("sqlite:" + filepath.Join(tmp, "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.(*memstore.SQLite); !ok {
		t.Fatalf("expected SQLite, got %T", s)
	}

	s, err = memstore.StoreFactory("file:" + filepath.Join(tmp, "data.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.(*memstore.File); !ok {
		t.Fatalf("expected File, got %T", s)
	}

	s, err = memstore.StoreFactory("mem:")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.(*memstore.InMemory); !ok {
		t.Fatalf("expected InMemory, got %T", s)
	}
}
