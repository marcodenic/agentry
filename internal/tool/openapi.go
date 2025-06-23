package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// FromOpenAPI parses an OpenAPI v3 spec and returns a Registry with one tool per operation.
func FromOpenAPI(data []byte) (Registry, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(loader.Context); err != nil {
		return nil, err
	}

	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = strings.TrimRight(doc.Servers[0].URL, "/")
	}

	reg := Registry{}
	for p, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			if op.OperationID == "" {
				continue
			}
			name := op.OperationID
			desc := op.Summary
			if desc == "" {
				desc = op.Description
			}
			schema := map[string]any{"type": "object", "properties": map[string]any{}}
			if op.RequestBody != nil && op.RequestBody.Value != nil {
				for ct, media := range op.RequestBody.Value.Content {
					if ct == "application/json" && media.Schema != nil && media.Schema.Value != nil {
						b, _ := json.Marshal(media.Schema.Value)
						_ = json.Unmarshal(b, &schema)
						break
					}
				}
			}
			// query params
			props, _ := schema["properties"].(map[string]any)
			if props == nil {
				props = map[string]any{}
				schema["properties"] = props
			}
			required := []string{}
			for _, paramRef := range op.Parameters {
				param := paramRef.Value
				if param == nil {
					continue
				}
				if param.In == "query" {
					if param.Schema != nil && param.Schema.Value != nil {
						b, _ := json.Marshal(param.Schema.Value)
						var m map[string]any
						_ = json.Unmarshal(b, &m)
						props[param.Name] = m
					} else {
						props[param.Name] = map[string]any{"type": "string"}
					}
					if param.Required {
						required = append(required, param.Name)
					}
				}
			}
			if len(required) > 0 {
				schema["required"] = required
			}

			endpoint := baseURL + path.Join("/", p)
			methodUpper := strings.ToUpper(method)

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
			}(endpoint, methodUpper)

			reg[name] = NewWithSchema(name, desc, schema, exec)
		}
	}
	return reg, nil
}
