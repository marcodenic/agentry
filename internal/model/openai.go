package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// OpenAI client uses OpenAI's chat completion API.
type OpenAI struct {
	key    string
	client *http.Client
}

func NewOpenAI(key string) *OpenAI {
	return &OpenAI{key: key, client: http.DefaultClient}
}

func (o *OpenAI) Complete(ctx context.Context, prompt string) (string, error) {
	if o.key == "" {
		return "", errors.New("missing api key")
	}

	reqBody := map[string]any{
		"model":       "gpt-4o",
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  64,
		"temperature": 0,
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.key)
	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(string(body))
	}
	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if len(res.Choices) == 0 {
		return "", errors.New("no choices")
	}
	return res.Choices[0].Message.Content, nil
}
