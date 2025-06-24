package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/memory"
)

func TestInMemoryVectorSimilarity(t *testing.T) {
	v := memory.NewInMemoryVector()
	v.Add(context.Background(), "a", "hello world")
	v.Add(context.Background(), "b", "goodbye moon")
	ids, err := v.Query(context.Background(), "hello", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "a" {
		t.Fatalf("expected a, got %#v", ids)
	}
}
