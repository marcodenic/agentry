package tests

import (
	"context"
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/model"
)

func TestOpenAIClient(t *testing.T) {
	key := os.Getenv("OPENAI_KEY")
	if key == "" {
		t.Skip("OPENAI_KEY not set")
	}
	c := model.NewOpenAI(key)
	out, err := c.Complete(context.Background(), "Hello")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Errorf("empty response")
	}
}
