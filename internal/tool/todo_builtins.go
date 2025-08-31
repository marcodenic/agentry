package tool

import (
    "context"
    "crypto/sha1"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/marcodenic/agentry/internal/memstore"
)

type todoItem struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description,omitempty"`
    Priority    string    `json:"priority,omitempty"` // low|med|high
    Tags        []string  `json:"tags,omitempty"`
    AgentID     string    `json:"agent_id,omitempty"`
    Status      string    `json:"status"` // pending|done
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Source      string    `json:"source,omitempty"` // optional file:line link
}

func todoNamespace() string {
    cwd, _ := os.Getwd()
    abs, _ := filepath.Abs(cwd)
    h := sha1.Sum([]byte(abs))
    return "todo:project:" + hex.EncodeToString(h[:8])
}

func todoKey(id string) string { return "item:" + id }

func putTodo(ns string, it todoItem) error {
    b, err := json.Marshal(it)
    if err != nil {
        return err
    }
    return memstore.Get().Set(ns, todoKey(it.ID), b, 0)
}

func getTodo(ns, id string) (todoItem, bool, error) {
    b, ok, err := memstore.Get().Get(ns, todoKey(id))
    if err != nil || !ok {
        return todoItem{}, false, err
    }
    var it todoItem
    if err := json.Unmarshal(b, &it); err != nil {
        return todoItem{}, false, err
    }
    return it, true, nil
}

func listTodos(ns string) ([]todoItem, error) {
    keys, err := memstore.Get().Keys(ns)
    if err != nil {
        return nil, err
    }
    var items []todoItem
    for _, k := range keys {
        if !strings.HasPrefix(k, "item:") {
            continue
        }
        b, ok, err := memstore.Get().Get(ns, k)
        if err != nil || !ok {
            continue
        }
        var it todoItem
        if json.Unmarshal(b, &it) == nil {
            items = append(items, it)
        }
    }
    sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
    return items, nil
}

func init() {
    // Add
    builtinMap["todo_add"] = builtinSpec{
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
            },
            "required": []string{"title"},
        },
        Exec: func(ctx context.Context, args map[string]any) (string, error) {
            title, _ := args["title"].(string)
            if strings.TrimSpace(title) == "" {
                return "", errors.New("title is required")
            }
            now := time.Now()
            // Generate simple sortable ID
            id := fmt.Sprintf("%d", now.UnixNano())
            it := todoItem{
                ID:          id,
                Title:       title,
                Description: strArg(args, "description"),
                Priority:    strArg(args, "priority"),
                Tags:        strSlice(args, "tags"),
                AgentID:     strArg(args, "agent_id"),
                Status:      "pending",
                CreatedAt:   now,
                UpdatedAt:   now,
                Source:      strArg(args, "source"),
            }
            ns := todoNamespace()
            if err := putTodo(ns, it); err != nil {
                return "", err
            }
            return marshal(map[string]any{"ok": true, "id": it.ID, "item": it})
        },
    }

    // List
    builtinMap["todo_list"] = builtinSpec{
        Desc: "List TODO items (filters: status, priority, tags)",
        Schema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "status":   map[string]any{"type": "string", "enum": []string{"pending", "done"}},
                "priority": map[string]any{"type": "string", "enum": []string{"low", "med", "high"}},
                "tags":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
            },
        },
        Exec: func(ctx context.Context, args map[string]any) (string, error) {
            ns := todoNamespace()
            items, err := listTodos(ns)
            if err != nil {
                return "", err
            }
            status := strings.TrimSpace(strArg(args, "status"))
            priority := strings.TrimSpace(strArg(args, "priority"))
            tags := strSlice(args, "tags")
            var out []todoItem
            next: for _, it := range items {
                if status != "" && it.Status != status { continue next }
                if priority != "" && it.Priority != priority { continue next }
                if len(tags) > 0 {
                    if !hasAllTags(it.Tags, tags) { continue next }
                }
                out = append(out, it)
            }
            return marshal(map[string]any{"ok": true, "count": len(out), "items": out})
        },
    }

    // Get
    builtinMap["todo_get"] = builtinSpec{
        Desc: "Get a TODO item by id",
        Schema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}},
        Exec: func(ctx context.Context, args map[string]any) (string, error) {
            id := strArg(args, "id")
            if id == "" { return "", errors.New("id is required") }
            ns := todoNamespace()
            it, ok, err := getTodo(ns, id)
            if err != nil { return "", err }
            if !ok { return marshal(map[string]any{"ok": false, "error": "not found"}) }
            return marshal(map[string]any{"ok": true, "item": it})
        },
    }

    // Update
    builtinMap["todo_update"] = builtinSpec{
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
        Exec: func(ctx context.Context, args map[string]any) (string, error) {
            id := strArg(args, "id")
            if id == "" { return "", errors.New("id is required") }
            ns := todoNamespace()
            it, ok, err := getTodo(ns, id)
            if err != nil { return "", err }
            if !ok { return "", errors.New("todo not found") }
            // Apply updates
            if v := strArg(args, "title"); v != "" { it.Title = v }
            if v := strArg(args, "description"); v != "" { it.Description = v }
            if v := strArg(args, "priority"); v != "" { it.Priority = v }
            if v := strArg(args, "agent_id"); v != "" { it.AgentID = v }
            if v := strArg(args, "status"); v != "" { it.Status = v }
            if _, ok := args["tags"].([]any); ok {
                it.Tags = strSlice(args, "tags")
            }
            it.UpdatedAt = time.Now()
            if err := putTodo(ns, it); err != nil { return "", err }
            return marshal(map[string]any{"ok": true, "item": it})
        },
    }

    // Delete
    builtinMap["todo_delete"] = builtinSpec{
        Desc: "Delete a TODO item by id",
        Schema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}},
        Exec: func(ctx context.Context, args map[string]any) (string, error) {
            id := strArg(args, "id")
            if id == "" { return "", errors.New("id is required") }
            ns := todoNamespace()
            if err := memstore.Get().Delete(ns, todoKey(id)); err != nil {
                return "", err
            }
            return marshal(map[string]any{"ok": true, "deleted": id})
        },
    }
}

func strArg(args map[string]any, key string) string {
    if v, ok := args[key].(string); ok { return v }
    return ""
}

func strSlice(args map[string]any, key string) []string {
    v, ok := args[key]
    if !ok { return nil }
    arr, ok := v.([]any)
    if !ok { return nil }
    out := make([]string, 0, len(arr))
    for _, it := range arr {
        if s, ok := it.(string); ok { out = append(out, s) }
    }
    return out
}

func hasAllTags(have, want []string) bool {
    if len(want) == 0 { return true }
    set := map[string]struct{}{}
    for _, t := range have { set[strings.ToLower(t)] = struct{}{} }
    for _, t := range want {
        if _, ok := set[strings.ToLower(t)]; !ok { return false }
    }
    return true
}

