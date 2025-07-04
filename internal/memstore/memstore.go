package memstore

import (
	"context"
)

// KV defines a simple bucketed key/value store.
type KV interface {
	Set(ctx context.Context, bucket, key string, val []byte) error
	Get(ctx context.Context, bucket, key string) ([]byte, error)
}

// InMemory is a simple in-memory implementation that does nothing.
type InMemory struct{}

func NewInMemory() *InMemory { return &InMemory{} }

func (s *InMemory) Set(ctx context.Context, bucket, key string, val []byte) error {
	// No-op implementation
	return nil
}

func (s *InMemory) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	// No-op implementation
	return nil, nil
}

// StoreFactory creates a minimal store (only in-memory)
func StoreFactory(uri string) (KV, error) {
	return NewInMemory(), nil
}
