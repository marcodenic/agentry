package tests

import (
	"context"
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/model"
)

func TestOpenAIClient(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	c := model.NewOpenAI(key, "gpt-4o")
	msgs := []model.ChatMessage{{Role: "user", Content: "Hello"}}
	comp, err := c.Complete(context.Background(), msgs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp.Content == "" {
		t.Errorf("empty response")
	}
}
