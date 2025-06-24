package memstore

import (
	"context"
	"time"
)

// KV defines a simple bucketed key/value store.
type KV interface {
	Set(ctx context.Context, bucket, key string, val []byte) error
	Get(ctx context.Context, bucket, key string) ([]byte, error)
}

// Vector defines storage for text documents retrievable via similarity.
type Vector interface {
	Add(ctx context.Context, id, text string) error
	Query(ctx context.Context, text string, k int) ([]string, error)
}

// Cleaner defines optional TTL-based cleanup for a store.
type Cleaner interface {
	Cleanup(ctx context.Context, bucket string, ttl time.Duration) error
}
