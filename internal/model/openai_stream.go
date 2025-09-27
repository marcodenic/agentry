package model

import (
	"context"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

type streamResult struct {
	content      string
	toolCalls    []ToolCall
	inputTokens  int
	outputTokens int
	modelName    string
	responseID   string
	chunkCount   int
	firstToken   bool
}

func collectStream(ctx context.Context, stream <-chan StreamChunk) (streamResult, error) {
	var res streamResult
	var sb strings.Builder
	start := time.Now()
	for chunk := range stream {
		res.chunkCount++
		debug.Printf("OpenAI.Stream: Received chunk %d", res.chunkCount)
		if chunk.Err != nil {
			debug.Printf("OpenAI.Stream: Chunk error: %v", chunk.Err)
			return res, chunk.Err
		}
		if chunk.ContentDelta != "" {
			sb.WriteString(chunk.ContentDelta)
			if !res.firstToken {
				res.firstToken = true
				debug.Printf("OpenAI.Stream: First token received")
			}
		}
		if chunk.Done {
			res.toolCalls = chunk.ToolCalls
			if chunk.InputTokens > 0 {
				res.inputTokens = chunk.InputTokens
			}
			if chunk.OutputTokens > 0 {
				res.outputTokens = chunk.OutputTokens
			}
			if chunk.ModelName != "" {
				res.modelName = chunk.ModelName
			}
			if chunk.ResponseID != "" {
				res.responseID = chunk.ResponseID
			}
		}
	}
	duration := time.Since(start)
	debug.Printf("OpenAI.Stream: Stream reading completed with %d chunks, read_duration=%v", res.chunkCount, duration)
	res.content = sb.String()
	return res, nil
}
