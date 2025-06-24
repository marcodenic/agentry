package memory

import (
	"context"
	"math"
	"sort"
	"strings"
)

// VectorStore defines minimal interface for vector retrieval.
type VectorStore interface {
	Add(ctx context.Context, id, text string) error
	Query(ctx context.Context, text string, k int) ([]string, error)
}

// Simple in-memory cosine-sim implementation for demo.

// InMemoryVector is a naive store keeping text docs.
type InMemoryVector struct {
	docs map[string]string
	vecs map[string]map[string]float64
}

func NewInMemoryVector() *InMemoryVector {
	return &InMemoryVector{docs: make(map[string]string), vecs: make(map[string]map[string]float64)}
}

func (v *InMemoryVector) Add(_ context.Context, id, text string) error {
	v.docs[id] = text
	v.vecs[id] = embed(text)
	return nil
}

func (v *InMemoryVector) Query(_ context.Context, text string, k int) ([]string, error) {
	qv := embed(text)
	type pair struct {
		id  string
		sim float64
	}
	list := make([]pair, 0, len(v.vecs))
	for id, vec := range v.vecs {
		list = append(list, pair{id: id, sim: cosine(vec, qv)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].sim > list[j].sim })
	if k > len(list) {
		k = len(list)
	}
	res := make([]string, 0, k)
	for i := 0; i < k; i++ {
		res = append(res, list[i].id)
	}
	return res, nil
}

func embed(text string) map[string]float64 {
	vec := map[string]float64{}
	for _, w := range strings.Fields(strings.ToLower(text)) {
		vec[w]++
	}
	return vec
}

func cosine(a, b map[string]float64) float64 {
	var dot, normA, normB float64
	for k, av := range a {
		dot += av * b[k]
		normA += av * av
	}
	for _, bv := range b {
		normB += bv * bv
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
