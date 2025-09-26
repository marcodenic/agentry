package model

import (
	"context"
	"encoding/json"
	"io"
	"testing"
)

func TestBuildRequestWithToolOutputs(t *testing.T) {
	client := NewOpenAI("test-key", "gpt-4o")
	client.previousResponseID = "resp_test_12345"

	msgs := []ChatMessage{
		{Role: "user", Content: "Test message"},
		{Role: "tool", ToolCallID: "call_test_123", Content: "Tool result"},
	}

	req, err := client.buildRequest(context.Background(), msgs, nil, true)
	if err != nil {
		t.Fatalf("buildRequest failed: %v", err)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	var bodyData map[string]any
	if err := json.Unmarshal(body, &bodyData); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	if req.URL.String() != "https://api.openai.com/v1/responses" {
		t.Fatalf("unexpected request URL: %s", req.URL.String())
	}

	if bodyData["previous_response_id"] != "resp_test_12345" {
		t.Fatalf("missing or incorrect previous_response_id: %v", bodyData["previous_response_id"])
	}

	rawInput, ok := bodyData["input"].([]any)
	if !ok {
		t.Fatalf("input should be an array: %#v", bodyData["input"])
	}

	if len(rawInput) != 1 {
		t.Fatalf("expected 1 input item, got %d", len(rawInput))
	}

	entry, ok := rawInput[0].(map[string]any)
	if !ok {
		t.Fatalf("input entry should be an object: %#v", rawInput[0])
	}

	if entry["type"] != "function_call_output" {
		t.Fatalf("expected type=function_call_output, got %v", entry["type"])
	}

	if entry["call_id"] != "call_test_123" {
		t.Fatalf("call_id mismatch: %v", entry["call_id"])
	}

	if entry["output"] != "Tool result" {
		t.Fatalf("output mismatch: %v", entry["output"])
	}

	if _, exists := bodyData["tool_outputs"]; exists {
		t.Fatalf("tool_outputs should not be present in continuation body: %#v", bodyData["tool_outputs"])
	}
}

func TestBuildRequestWithoutToolOutputs(t *testing.T) {
	client := NewOpenAI("test-key", "gpt-4o")

	msgs := []ChatMessage{
		{Role: "user", Content: "Test message"},
	}

	req, err := client.buildRequest(context.Background(), msgs, nil, true)
	if err != nil {
		t.Fatalf("buildRequest failed: %v", err)
	}

	if req.URL.String() != "https://api.openai.com/v1/responses" {
		t.Fatalf("unexpected request URL: %s", req.URL.String())
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	var bodyData map[string]any
	if err := json.Unmarshal(body, &bodyData); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	rawInput, ok := bodyData["input"].([]any)
	if !ok {
		t.Fatalf("input should be an array: %#v", bodyData["input"])
	}

	if len(rawInput) != 1 {
		t.Fatalf("expected 1 input item, got %d", len(rawInput))
	}

	entry, ok := rawInput[0].(map[string]any)
	if !ok {
		t.Fatalf("input entry should be an object: %#v", rawInput[0])
	}

	if entry["role"] != "user" {
		t.Fatalf("expected role=user, got %v", entry["role"])
	}

	content, ok := entry["content"].([]any)
	if !ok {
		t.Fatalf("content should be an array: %#v", entry["content"])
	}

	if len(content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(content))
	}

	part, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("content part should be an object: %#v", content[0])
	}

	if part["type"] != "input_text" {
		t.Fatalf("expected type=input_text, got %v", part["type"])
	}

	if part["text"] != "Test message" {
		t.Fatalf("expected text=Test message, got %v", part["text"])
	}

	if _, exists := bodyData["previous_response_id"]; exists {
		t.Fatalf("previous_response_id should be absent for new turn: %#v", bodyData["previous_response_id"])
	}
}

func TestBuildRequestSkipsStaleToolOutputs(t *testing.T) {
	client := NewOpenAI("test-key", "gpt-4o")
	client.previousResponseID = "resp_test_67890"

	msgs := []ChatMessage{
		{Role: "user", Content: "Test message"},
		{Role: "assistant", Content: "", ToolCalls: []ToolCall{{ID: "call_test_old"}}},
		{Role: "tool", ToolCallID: "call_test_old", Content: "Old result"},
		{Role: "assistant", Content: "", ToolCalls: []ToolCall{{ID: "call_test_new"}}},
		{Role: "tool", ToolCallID: "call_test_new", Content: "New result"},
	}

	req, err := client.buildRequest(context.Background(), msgs, nil, true)
	if err != nil {
		t.Fatalf("buildRequest failed: %v", err)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}

	var bodyData map[string]any
	if err := json.Unmarshal(body, &bodyData); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}

	rawInput, ok := bodyData["input"].([]any)
	if !ok {
		t.Fatalf("input should be an array: %#v", bodyData["input"])
	}

	if len(rawInput) != 1 {
		t.Fatalf("expected 1 input item, got %d", len(rawInput))
	}

	entry, ok := rawInput[0].(map[string]any)
	if !ok {
		t.Fatalf("input entry should be an object: %#v", rawInput[0])
	}

	if entry["call_id"] != "call_test_new" {
		t.Fatalf("expected only latest call_id, got %v", entry["call_id"])
	}

	if entry["output"] != "New result" {
		t.Fatalf("output mismatch: %v", entry["output"])
	}
}
