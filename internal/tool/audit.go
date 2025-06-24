package tool

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

// AuditEvent represents a tool execution event.
type AuditEvent struct {
	Tool      string         `json:"tool"`
	Args      map[string]any `json:"args"`
	Duration  int64          `json:"duration_ms"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"ts"`
}

// WrapWithAudit wraps all tools in a registry with audit logging to w.
func WrapWithAudit(reg Registry, w io.Writer) Registry {
	out := Registry{}
	for name, t := range reg {
		out[name] = auditTool{Tool: t, w: w}
	}
	return out
}

type auditTool struct {
	Tool
	w io.Writer
}

func (a auditTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	start := time.Now()
	res, err := a.Tool.Execute(ctx, args)
	evt := AuditEvent{
		Tool:      a.Name(),
		Args:      args,
		Duration:  time.Since(start).Milliseconds(),
		Timestamp: time.Now().UTC(),
	}
	if err != nil {
		evt.Error = err.Error()
	}
	b, _ := json.Marshal(evt)
	_, _ = wWrite(a.w, b)
	return res, err
}

func wWrite(w io.Writer, b []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	n, err := w.Write(append(b, '\n'))
	return n, err
}
