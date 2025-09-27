package model

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

type OpenAIConversation struct {
	client *OpenAI
	msgs   []ChatMessage
	tools  []ToolSpec
	start  time.Time
}

func newOpenAIConversation(client *OpenAI, msgs []ChatMessage, tools []ToolSpec) *OpenAIConversation {
	return &OpenAIConversation{client: client, msgs: msgs, tools: tools}
}

func (c *OpenAIConversation) Stream(ctx context.Context) (<-chan StreamChunk, error) {
	c.start = time.Now()
	debug.Printf("OpenAI.Stream: START msgs=%d tools=%d", len(c.msgs), len(c.tools))

	debug.LogModelInteraction("openai", c.client.model, len(c.msgs), map[string]int{"input_estimated": estimateTokens(c.msgs)}, 0)

	req, err := c.buildRequest(ctx)
	debug.Printf("OpenAI.Stream: buildRequest err=%v elapsed=%v", err, time.Since(c.start))
	if err != nil {
		debug.Printf("OpenAI.Stream: buildRequest failed: %v", err)
		debug.LogEvent("MODEL", "request_build_failed", map[string]interface{}{
			"provider": "openai",
			"model":    c.client.model,
			"error":    err.Error(),
		})
		return nil, err
	}

	if deadline, ok := ctx.Deadline(); ok {
		timeoutDur := time.Until(deadline)
		debug.Printf("OpenAI.Stream: context timeout set to %v", timeoutDur)
	} else {
		debug.Printf("OpenAI.Stream: no context timeout set")
	}

	debug.Printf("OpenAI.Stream: starting HTTP request goroutine, elapsed=%v", time.Since(c.start))
	out := make(chan StreamChunk, 32)
	go c.run(ctx, out, req)
	return out, nil
}

func (c *OpenAIConversation) buildRequest(ctx context.Context) (*http.Request, error) {
	builder := newOARequestBuilder(c.client, c.msgs, c.tools)
	return builder.Build(ctx, true)
}

func (c *OpenAIConversation) run(ctx context.Context, out chan<- StreamChunk, req *http.Request) {
	defer close(out)
	reqStartTime := time.Now()
	debug.Printf("OpenAI.Stream: HTTP request starting... elapsed=%v", time.Since(c.start))

	resp, err := c.client.client.Do(req)
	reqEndTime := time.Now()
	debug.Printf("OpenAI.Stream: HTTP response received err=%v request_duration=%v total_elapsed=%v", err, reqEndTime.Sub(reqStartTime), reqEndTime.Sub(c.start))

	if err != nil {
		debug.Printf("OpenAI.Stream: HTTP request failed: %v", err)
		debug.LogEvent("MODEL", "http_request_failed", map[string]interface{}{
			"provider": "openai",
			"model":    c.client.model,
			"error":    err.Error(),
			"duration": reqEndTime.Sub(reqStartTime).String(),
		})
		out <- StreamChunk{Err: err}
		return
	}
	defer resp.Body.Close()
	debug.Printf("OpenAI.Stream: status=%d headers=%v", resp.StatusCode, resp.Header)
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		debug.LogEvent("MODEL", "api_error", map[string]interface{}{
			"provider": "openai",
			"model":    c.client.model,
			"status":   resp.StatusCode,
			"response": string(body),
		})
		out <- StreamChunk{Err: errors.New(string(body))}
		return
	}

	reader := c.newStreamReader()
	result, readErr := reader.Read(ctx, resp.Body, func(chunk StreamChunk) {
		out <- chunk
	})
	if readErr != nil {
		return
	}
	if result == nil {
		return
	}
	if id := result.ResponseID(); id != "" {
		c.client.previousResponseID = id
		debug.Printf("OpenAI.Stream: Persisting response ID for next request: %s", id)
	}
	result.Finalize(out)
}

func (c *OpenAIConversation) newStreamReader() streamReader {
	return newOpenAIStreamReader(c.client.model, len(c.msgs), c.start)
}
