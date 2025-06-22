package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Qdrant implements VectorStore against the Qdrant REST API.
type Qdrant struct {
	endpoint   string
	collection string
	client     *http.Client
}

// NewQdrant returns a new Qdrant store pointing at the given endpoint and collection.
func NewQdrant(endpoint, collection string) *Qdrant {
	return &Qdrant{endpoint: endpoint, collection: collection, client: &http.Client{}}
}

func (q *Qdrant) Add(ctx context.Context, id, text string) error {
	payload := map[string]any{
		"points": []map[string]any{
			{
				"id":      id,
				"vector":  []float32{0},
				"payload": map[string]string{"text": text},
			},
		},
	}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/collections/%s/points", q.endpoint, q.collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("qdrant: %s", resp.Status)
	}
	return nil
}

func (q *Qdrant) Query(ctx context.Context, text string, k int) ([]string, error) {
	payload := map[string]any{
		"vector": []float32{0},
		"limit":  k,
	}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/collections/%s/points/search", q.endpoint, q.collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("qdrant: %s", resp.Status)
	}
	var out struct {
		Result []struct {
			ID string `json:"id"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(out.Result))
	for _, r := range out.Result {
		ids = append(ids, r.ID)
	}
	return ids, nil
}
