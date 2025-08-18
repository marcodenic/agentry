package tokens

import (
	"fmt"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

var (
	// Cache encoders to avoid repeated initialization
	encoderCache = make(map[string]*tiktoken.Tiktoken)
	cacheMutex   sync.RWMutex
)

// getEncoder returns a cached tiktoken encoder for the given model
func getEncoder(modelName string) (*tiktoken.Tiktoken, error) {
	cacheMutex.RLock()
	if encoder, exists := encoderCache[modelName]; exists {
		cacheMutex.RUnlock()
		return encoder, nil
	}
	cacheMutex.RUnlock()

	// Determine encoding based on model
	var encoding string
	switch {
	case contains(modelName, "gpt-4"), contains(modelName, "gpt-3.5"), contains(modelName, "text-embedding"):
		encoding = "cl100k_base"
	case contains(modelName, "gpt-3"), contains(modelName, "davinci"), contains(modelName, "curie"), contains(modelName, "babbage"), contains(modelName, "ada"):
		encoding = "p50k_base"
	case contains(modelName, "claude"):
		// Claude uses a different tokenizer, but cl100k_base is a reasonable approximation
		encoding = "cl100k_base"
	default:
		// Default to cl100k_base for unknown models (most modern models use this)
		encoding = "cl100k_base"
	}

	encoder, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding %s: %w", encoding, err)
	}

	cacheMutex.Lock()
	encoderCache[modelName] = encoder
	cacheMutex.Unlock()

	return encoder, nil
}

// contains is a helper function for case-insensitive substring matching
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Count returns the number of tokens in the given text for the specified model
func Count(text, modelName string) int {
	if text == "" {
		return 0
	}

	encoder, err := getEncoder(modelName)
	if err != nil {
		// Fallback to word-based estimation if tiktoken fails
		return fallbackWordCount(text)
	}

	tokens := encoder.Encode(text, nil, nil)
	return len(tokens)
}

// CountWithFallback counts tokens but falls back to a simple heuristic if tiktoken fails
func CountWithFallback(text string) int {
	if text == "" {
		return 0
	}

	// Try with a default modern model encoding
	encoder, err := getEncoder("gpt-4")
	if err != nil {
		return fallbackWordCount(text)
	}

	tokens := encoder.Encode(text, nil, nil)
	return len(tokens)
}

// fallbackWordCount provides a simple word-based token estimation
func fallbackWordCount(text string) int {
	if text == "" {
		return 0
	}
	// Use word count * 1.3 as a reasonable approximation
	// (tokens are typically 75% of word count for English text)
	words := len(splitFields(text))
	return int(float64(words) * 1.3)
}

// splitFields is a simple field splitter that mimics strings.Fields behavior
func splitFields(s string) []string {
	var fields []string
	var current []rune

	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if len(current) > 0 {
				fields = append(fields, string(current))
				current = current[:0]
			}
		} else {
			current = append(current, r)
		}
	}

	if len(current) > 0 {
		fields = append(fields, string(current))
	}

	return fields
}

// Truncate truncates text to fit within the specified token limit for the given model
func Truncate(text string, tokenLimit int, modelName string) string {
	if text == "" {
		return text
	}

	currentTokens := Count(text, modelName)
	if currentTokens <= tokenLimit {
		return text
	}

	// Binary search to find the right truncation point
	left, right := 0, len(text)
	for left < right {
		mid := (left + right + 1) / 2
		truncated := text[:mid]
		if Count(truncated, modelName) <= tokenLimit-10 { // Leave room for "...[truncated]"
			left = mid
		} else {
			right = mid - 1
		}
	}

	if left == 0 {
		return "...[content too large]..."
	}

	truncated := text[:left]
	// Try to break at a reasonable boundary
	if lastNewline := findLastIndex(truncated, '\n'); lastNewline > len(truncated)/2 {
		truncated = truncated[:lastNewline]
	}

	return truncated + "\n...[truncated]..."
}

// findLastIndex finds the last occurrence of a character in a string
func findLastIndex(s string, char rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if rune(s[i]) == char {
			return i
		}
	}
	return -1
}
