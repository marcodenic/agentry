package tool

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// getUtilityBuiltins returns utility builtin tools like echo and ping
func getUtilityBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"echo": {
			Desc: "Repeat a string",
			Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"text": map[string]any{"type": "string"}},
				"required":   []string{"text"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				txt, _ := args["text"].(string)
				return txt, nil
			},
		},
		"ping": {
			Desc: "Ping a host",
			Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"host": map[string]any{"type": "string"}},
				"required":   []string{"host"},
				"example":    map[string]any{"host": "example.com"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				host, _ := args["host"].(string)
				if host == "" {
					return "", errors.New("missing host")
				}
				d := net.Dialer{Timeout: 3 * time.Second}
				start := time.Now()
				conn, err := d.DialContext(ctx, "tcp", host+":80")
				if err != nil {
					return "", err
				}
				_ = conn.Close()
				return fmt.Sprintf("pong in %v", time.Since(start)), nil
			},
		},
	}
}
