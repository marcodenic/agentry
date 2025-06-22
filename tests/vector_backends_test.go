package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcodenic/agentry/internal/memory"
)

func TestQdrantAdapter(t *testing.T) {
	stored := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/collections/test/points":
			var req struct {
				Points []struct {
					ID      string            `json:"id"`
					Payload map[string]string `json:"payload"`
				} `json:"points"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Error(err)
				return
			}
			for _, p := range req.Points {
				stored[p.ID] = p.Payload["text"]
			}
			w.WriteHeader(200)
		case r.Method == http.MethodPost && r.URL.Path == "/collections/test/points/search":
			ids := []struct {
				ID string `json:"id"`
			}{}
			i := 0
			for id := range stored {
				if i >= 1 {
					break
				}
				ids = append(ids, struct {
					ID string `json:"id"`
				}{ID: id})
				i++
			}
			json.NewEncoder(w).Encode(map[string]any{"result": ids})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	q := memory.NewQdrant(srv.URL, "test")
	if err := q.Add(context.Background(), "a", "hello"); err != nil {
		t.Fatal(err)
	}
	ids, err := q.Query(context.Background(), "hello", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "a" {
		t.Fatalf("unexpected ids: %#v", ids)
	}
}

func TestFaissAdapter(t *testing.T) {
	stored := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/add":
			var req struct{ ID, Text string }
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Error(err)
				return
			}
			stored[req.ID] = req.Text
		case "/query":
			var req struct {
				Text string
				K    int
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Error(err)
				return
			}
			ids := []string{}
			i := 0
			for id := range stored {
				if i >= req.K {
					break
				}
				ids = append(ids, id)
				i++
			}
			json.NewEncoder(w).Encode(ids)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	f := memory.NewFaiss(srv.URL)
	if err := f.Add(context.Background(), "x", "hello"); err != nil {
		t.Fatal(err)
	}
	ids, err := f.Query(context.Background(), "hello", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "x" {
		t.Fatalf("unexpected ids: %#v", ids)
	}
}
