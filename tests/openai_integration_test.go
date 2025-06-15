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
	msgs := []model.ChatMessage{{Role: "user", Content: "Hello"}}
	comp, err := c.Complete(context.Background(), msgs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp.Content == "" {
		t.Errorf("empty response")
	}
}
