package tool

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/marcodenic/agentry/internal/todo"
)

var (
	todoOnce sync.Once
	todoSvc  *todo.Service
	todoErr  error
)

func todoTools() map[string]builtinSpec {
	return map[string]builtinSpec{
		"todo_add":    todoAddSpec(),
		"todo_list":   todoListSpec(),
		"todo_get":    todoGetSpec(),
		"todo_update": todoUpdateSpec(),
		"todo_delete": todoDeleteSpec(),
	}
}

func getTodoBuiltins() map[string]builtinSpec {
	return todoTools()
}

func registerTodoBuiltins(reg *builtinRegistry) {
	reg.addAll(getTodoBuiltins())
}

func acquireTodoService() (*todo.Service, error) {
	todoOnce.Do(func() {
		todoSvc, todoErr = todo.NewService(nil, "")
	})
	return todoSvc, todoErr
}

func todoAddSpec() builtinSpec {
	return builtinSpec{
		Desc: "Add a TODO item to the project planning list",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":       map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
				"priority":    map[string]any{"type": "string", "enum": []string{"low", "med", "high"}},
				"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"agent_id":    map[string]any{"type": "string"},
				"source":      map[string]any{"type": "string"},
				"status":      map[string]any{"type": "string"},
			},
			"required": []string{"title"},
		},
		Exec: todoAddExec,
	}
}

func todoAddExec(ctx context.Context, args map[string]any) (string, error) {
	title := stringArg(args, "title")
	if strings.TrimSpace(title) == "" {
		return "", errors.New("title is required")
	}

	svc, err := acquireTodoService()
	if err != nil {
		return "", err
	}

	item, err := svc.Create(todo.CreateParams{
		Title:       title,
		Description: strArg(args, "description"),
		Priority:    strArg(args, "priority"),
		Tags:        strSlice(args, "tags"),
		AgentID:     strArg(args, "agent_id"),
		Source:      strArg(args, "source"),
		Status:      strArg(args, "status"),
	})
	if err != nil {
		return "", err
	}

	return marshal(map[string]any{"ok": true, "id": item.ID, "item": item})
}

func todoListSpec() builtinSpec {
	return builtinSpec{
		Desc: "List TODO items (filters: status, priority, tags)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status":   map[string]any{"type": "string", "enum": []string{"pending", "done"}},
				"priority": map[string]any{"type": "string", "enum": []string{"low", "med", "high"}},
				"tags":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
		},
		Exec: todoListExec,
	}
}

func todoListExec(ctx context.Context, args map[string]any) (string, error) {
	svc, err := acquireTodoService()
	if err != nil {
		return "", err
	}

	filter := todo.Filter{
		Status:   strArg(args, "status"),
		Priority: strArg(args, "priority"),
		Tags:     strSlice(args, "tags"),
	}

	items, err := svc.List(filter)
	if err != nil {
		return "", err
	}

	return marshal(map[string]any{"ok": true, "count": len(items), "items": items})
}

func todoGetSpec() builtinSpec {
	return builtinSpec{
		Desc: "Get a TODO item by id",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string"}},
			"required":   []string{"id"},
		},
		Exec: todoGetExec,
	}
}

func todoGetExec(ctx context.Context, args map[string]any) (string, error) {
	id := strArg(args, "id")
	if id == "" {
		return "", errors.New("id is required")
	}

	svc, err := acquireTodoService()
	if err != nil {
		return "", err
	}

	item, ok, err := svc.Get(id)
	if err != nil {
		return "", err
	}
	if !ok {
		return marshal(map[string]any{"ok": false, "error": "not found"})
	}
	return marshal(map[string]any{"ok": true, "item": item})
}

func todoUpdateSpec() builtinSpec {
	return builtinSpec{
		Desc: "Update a TODO item (any field)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":          map[string]any{"type": "string"},
				"title":       map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
				"priority":    map[string]any{"type": "string", "enum": []string{"low", "med", "high"}},
				"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"agent_id":    map[string]any{"type": "string"},
				"status":      map[string]any{"type": "string", "enum": []string{"pending", "done"}},
			},
			"required": []string{"id"},
		},
		Exec: todoUpdateExec,
	}
}

func todoUpdateExec(ctx context.Context, args map[string]any) (string, error) {
	id := strArg(args, "id")
	if id == "" {
		return "", errors.New("id is required")
	}

	svc, err := acquireTodoService()
	if err != nil {
		return "", err
	}

	updates := todo.UpdateParams{}
	if v := strArg(args, "title"); v != "" {
		updates.Title = ptr(v)
	}
	if v := strArg(args, "description"); v != "" {
		updates.Description = ptr(v)
	}
	if v := strArg(args, "priority"); v != "" {
		updates.Priority = ptr(v)
	}
	if v := strArg(args, "agent_id"); v != "" {
		updates.AgentID = ptr(v)
	}
	if v := strArg(args, "status"); v != "" {
		updates.Status = ptr(v)
	}
	if raw, ok := args["tags"].([]any); ok {
		tags := strSlice(map[string]any{"tags": raw}, "tags")
		updates.Tags = &tags
	}

	item, err := svc.Update(id, updates)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			return "", errors.New("todo not found")
		}
		return "", err
	}

	return marshal(map[string]any{"ok": true, "item": item})
}

func todoDeleteSpec() builtinSpec {
	return builtinSpec{
		Desc: "Delete a TODO item by id",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"id": map[string]any{"type": "string"}},
			"required":   []string{"id"},
		},
		Exec: todoDeleteExec,
	}
}

func todoDeleteExec(ctx context.Context, args map[string]any) (string, error) {
	id := strArg(args, "id")
	if id == "" {
		return "", errors.New("id is required")
	}

	svc, err := acquireTodoService()
	if err != nil {
		return "", err
	}

	if err := svc.Delete(id); err != nil {
		return "", err
	}
	return marshal(map[string]any{"ok": true, "deleted": id})
}

func strArg(args map[string]any, key string) string {
	return stringArg(args, key)
}

func strSlice(args map[string]any, key string) []string {
	v, ok := args[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		if s, ok := it.(string); ok {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}
