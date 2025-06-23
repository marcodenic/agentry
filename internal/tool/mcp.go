package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
)

type mcpCommand struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Method      string         `json:"method"`
	Path        string         `json:"path"`
	Schema      map[string]any `json:"schema"`
}

type mcpSpec struct {
	Server   string       `json:"server"`
	Commands []mcpCommand `json:"commands"`
}

// FromMCP parses a minimal MCP schema into a Registry.
func FromMCP(data []byte) (Registry, error) {
	var spec mcpSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	reg := Registry{}
	base := spec.Server
	for _, cmd := range spec.Commands {
		endpoint := base + path.Join("/", cmd.Path)
		method := http.MethodPost
		if cmd.Method != "" {
			method = cmd.Method
		}
		schema := cmd.Schema
		if schema == nil {
			schema = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		exec := func(endpoint, method string) ExecFn {
			return func(ctx context.Context, args map[string]any) (string, error) {
				b, err := json.Marshal(args)
				if err != nil {
					return "", err
				}
				req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(b))
				if err != nil {
					return "", err
				}
				if method != http.MethodGet {
					req.Header.Set("Content-Type", "application/json")
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return "", err
				}
				defer resp.Body.Close()
				rb, err := io.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				return string(rb), nil
			}
		}(endpoint, method)
		reg[cmd.Name] = NewWithSchema(cmd.Name, cmd.Description, schema, exec)
	}
	return reg, nil
}
