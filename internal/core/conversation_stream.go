package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/trace"
)

type streamExecutor struct {
	session *conversationSession
}

func newStreamExecutor(session *conversationSession) *streamExecutor {
	return &streamExecutor{session: session}
}

func (e *streamExecutor) Execute() (model.Completion, string, error) {
	s := e.session
	s.tracer.StreamInvocation(s.msgs)

	streamCh, err := e.invokeClient()
	if err != nil {
		return model.Completion{}, "", err
	}
	if streamCh == nil {
		return model.Completion{}, "", fmt.Errorf("streaming client returned nil channel")
	}

	aggregator := newChunkAggregator(s)
	if err := aggregator.Collect(streamCh); err != nil {
		return model.Completion{}, "", err
	}

	completion, responseID := aggregator.Result()
	s.tracer.CompletionReady(completion)
	s.agent.Trace(s.ctx, trace.EventStepStart, completion)
	if responseID != "" {
		s.agent.Trace(s.ctx, trace.EventSummary, map[string]any{"response_id": responseID})
	}

	return completion, responseID, nil
}

func (e *streamExecutor) invokeClient() (<-chan model.StreamChunk, error) {
	start := time.Now()
	ch, err := e.session.agent.Client.Stream(e.session.ctx, e.session.msgs, e.session.specs)
	debug.Printf("Agent.Run: MODEL CLIENT RETURNED - AFTER STREAM, err=%v, call_duration=%v", err, time.Since(start))
	return ch, err
}

type chunkAggregator struct {
	session            *conversationSession
	assembled          strings.Builder
	toolCalls          []model.ToolCall
	inputTokensUsed    int
	outputTokensUsed   int
	modelNameUsed      string
	responseIDUsed     string
	firstTokenRecorded bool
	chunkCount         int
}

func newChunkAggregator(session *conversationSession) *chunkAggregator {
	return &chunkAggregator{session: session}
}

func (a *chunkAggregator) Collect(stream <-chan model.StreamChunk) error {
	readStart := time.Now()
	for chunk := range stream {
		a.chunkCount++
		a.session.tracer.StreamChunk(a.chunkCount)
		if err := a.handleChunk(chunk); err != nil {
			return err
		}
	}
	a.session.tracer.StreamCompleted(a.chunkCount, time.Since(readStart))
	debug.Printf("Agent.Run: Assembled response length: %d chars", a.assembled.Len())
	return nil
}

func (a *chunkAggregator) handleChunk(chunk model.StreamChunk) error {
	if chunk.Err != nil {
		debug.Printf("Agent.Run: Chunk error: %v", chunk.Err)
		return chunk.Err
	}
	if delta := chunk.ContentDelta; delta != "" {
		a.assembled.WriteString(delta)
		if !a.firstTokenRecorded {
			a.firstTokenRecorded = true
			debug.Printf("Agent.Run: First token received")
		}
		a.session.agent.Trace(a.session.ctx, trace.EventToken, delta)
	}
	if chunk.Done {
		a.toolCalls = chunk.ToolCalls
		if chunk.InputTokens > 0 {
			a.inputTokensUsed = chunk.InputTokens
		}
		if chunk.OutputTokens > 0 {
			a.outputTokensUsed = chunk.OutputTokens
		}
		if chunk.ModelName != "" {
			a.modelNameUsed = chunk.ModelName
		}
		if chunk.ResponseID != "" {
			a.responseIDUsed = chunk.ResponseID
		}
	}
	return nil
}

func (a *chunkAggregator) Result() (model.Completion, string) {
	agent := a.session.agent
	completion := model.Completion{
		Content:      a.assembled.String(),
		ToolCalls:    a.toolCalls,
		InputTokens:  a.inputTokensUsed,
		OutputTokens: a.outputTokensUsed,
		ModelName: func() string {
			if strings.TrimSpace(a.modelNameUsed) != "" {
				return a.modelNameUsed
			}
			return agent.ModelName
		}(),
	}
	return completion, a.responseIDUsed
}
