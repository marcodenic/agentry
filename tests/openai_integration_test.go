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
	streamCh, err := c.Stream(context.Background(), msgs, nil)
	if err != nil {
		t.Fatal(err)
	}
	
	var content string
	for chunk := range streamCh {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		content += chunk.ContentDelta
		if chunk.Done {
			break
		}
	}
	
	if content == "" {
		t.Errorf("empty response")
	}
}
