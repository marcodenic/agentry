package tests

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/model"
)

func TestOpenAIClient(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	if conn, err := net.DialTimeout("tcp", "api.openai.com:443", time.Second); err != nil {
		if strings.Contains(err.Error(), "not permitted") || strings.Contains(err.Error(), "network is unreachable") {
			t.Skipf("network access unavailable: %v", err)
		}
	} else {
		conn.Close()
	}
	c := model.NewOpenAI(key, "gpt-4o")
	msgs := []model.ChatMessage{{Role: "user", Content: "Hello"}}
	stream, err := c.Stream(context.Background(), msgs, nil)
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for chunk := range stream {
		if chunk.Err != nil {
			t.Fatal(chunk.Err)
		}
		sb.WriteString(chunk.ContentDelta)
	}
	if sb.Len() == 0 {
		t.Errorf("empty response")
	}
}
