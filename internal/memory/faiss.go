package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Faiss implements VectorStore against a simple REST wrapper.
type Faiss struct {
	endpoint string
	client   *http.Client
}

// NewFaiss returns a new Faiss store.
func NewFaiss(endpoint string) *Faiss {
	return &Faiss{endpoint: endpoint, client: &http.Client{}}
}

func (f *Faiss) Add(ctx context.Context, id, text string) error {
	payload := map[string]string{"id": id, "text": text}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.endpoint+"/add", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("faiss: %s", resp.Status)
	}
	return nil
}

func (f *Faiss) Query(ctx context.Context, text string, k int) ([]string, error) {
	payload := map[string]any{"text": text, "k": k}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.endpoint+"/query", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("faiss: %s", resp.Status)
	}
	var out []string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
