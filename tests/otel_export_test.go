package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coltrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/marcodenic/agentry/internal/trace"
)

func TestOTLPExport(t *testing.T) {
	recv := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		var req coltrace.ExportTraceServiceRequest
		if err := proto.Unmarshal(body, &req); err != nil {
			t.Errorf("unmarshal: %v", err)
			return
		}
		if len(req.ResourceSpans) == 0 {
			t.Errorf("no spans")
		}
		recv <- struct{}{}
	}))
	defer srv.Close()

	shutdown, err := trace.Init(srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	w := trace.NewOTel()
	w.Write(context.Background(), trace.Event{Type: trace.EventFinal, AgentID: "a", Data: "done"})
	_ = shutdown(context.Background())

	select {
	case <-recv:
	case <-time.After(time.Second):
		t.Fatal("no request")
	}
}
