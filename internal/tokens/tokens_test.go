package tokens

import (
	"strings"
	"testing"

	"github.com/pkoukk/tiktoken-go"
)

func TestCountCachesEncoder(t *testing.T) {
	cacheMutex.Lock()
	encoderCache = make(map[string]*tiktoken.Tiktoken)
	cacheMutex.Unlock()

	text := "Measuring token counts is useful for budgeting."
	count := Count(text, "openai/gpt-4")
	if count <= 0 {
		t.Fatalf("expected positive token count, got %d", count)
	}

	cacheMutex.RLock()
	cached, ok := encoderCache["openai/gpt-4"]
	cacheMutex.RUnlock()
	if !ok || cached == nil {
		t.Fatalf("expected encoder cached for model")
	}

	second := Count(text, "openai/gpt-4")
	if second != count {
		t.Fatalf("expected deterministic count, got %d and %d", count, second)
	}
}

func TestTruncateAddsEllipsis(t *testing.T) {
	text := strings.Repeat("long input needs trimming ", 20)
	truncated := Truncate(text, 40, "openai/gpt-4")
	if len(truncated) >= len(text) {
		t.Fatalf("expected truncated string shorter than input")
	}
	if !strings.Contains(truncated, "...[truncated]...") {
		t.Fatalf("expected truncation marker, got %q", truncated)
	}
}

func TestFallbackWordCount(t *testing.T) {
	if n := fallbackWordCount(""); n != 0 {
		t.Fatalf("expected zero tokens for empty string, got %d", n)
	}
	if n := fallbackWordCount("two words"); n != 2 {
		t.Fatalf("expected approx two tokens, got %d", n)
	}
	words := splitFields("one\ttwo\nthree")
	if len(words) != 3 {
		t.Fatalf("expected splitFields to find three words, got %d", len(words))
	}
}
