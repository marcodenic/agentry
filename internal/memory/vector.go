package memory

import "context"

// VectorStore defines minimal interface for vector retrieval.
type VectorStore interface {
	Add(ctx context.Context, id, text string) error
	Query(ctx context.Context, text string, k int) ([]string, error)
}

// Simple in-memory cosine-sim implementation for demo.

// InMemoryVector is a naive store keeping text docs.
type InMemoryVector struct {
	docs map[string]string
}

func NewInMemoryVector() *InMemoryVector {
	return &InMemoryVector{docs: make(map[string]string)}
}

func (v *InMemoryVector) Add(_ context.Context, id, text string) error {
	v.docs[id] = text
	return nil
}

func (v *InMemoryVector) Query(_ context.Context, text string, k int) ([]string, error) {
	// Return first k IDs for now.
	res := make([]string, 0, k)
	for id := range v.docs {
		res = append(res, id)
		if len(res) >= k {
			break
		}
	}
	return res, nil
}
