package team

import (
	"encoding/json"
	"fmt"
	"time"

	teamruntime "github.com/marcodenic/agentry/internal/teamruntime"
)

// SetSharedData stores data in shared memory accessible to all agents
func (t *Team) SetSharedData(key string, value interface{}) {
	t.mutex.Lock()
	t.sharedMemory[key] = value
	t.mutex.Unlock()

	// Persist a JSON representation to the shared store (best-effort)
	if t.store != nil {
		if b, err := json.Marshal(value); err == nil {
			_ = t.store.Set(t.name, key, b, 0)
		} else {
			// Fallback to string formatting to avoid losing data entirely
			_ = t.store.Set(t.name, key, []byte(fmt.Sprintf("%v", value)), 0)
		}
	}

	// Log the shared memory update (append coordination event only in memory)
	event := CoordinationEvent{
		ID:        fmt.Sprintf("shared_%d", time.Now().UnixNano()),
		Type:      "shared_memory_update",
		From:      "system",
		To:        "*",
		Content:   fmt.Sprintf("Updated shared data: %s", key),
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"key": key, "value_type": fmt.Sprintf("%T", value)},
	}
	t.mutex.Lock()
	t.coordination = append(t.coordination, event)
	t.mutex.Unlock()
	teamruntime.Debugf("📊 Shared memory updated: %s\n", key)
}

// GetSharedData retrieves data from shared memory
func (t *Team) GetSharedData(key string) (interface{}, bool) {
	t.mutex.RLock()
	value, exists := t.sharedMemory[key]
	t.mutex.RUnlock()
	if exists {
		return value, true
	}

	// Try backing store if not present in in-memory map
	if t.store != nil {
		if b, ok, err := t.store.Get(t.name, key); err == nil && ok {
			var out interface{}
			if err := json.Unmarshal(b, &out); err != nil {
				// treat as plain string
				out = string(b)
			}
			// normalize common JSON generic types into typed Go forms used by callers
			out = normalizeShared(out)
			// cache in memory for quick typed access in-session
			t.mutex.Lock()
			t.sharedMemory[key] = out
			t.mutex.Unlock()
			return out, true
		}
	}

	return nil, false
}

// normalizeShared converts generic []any of map[string]any into []map[string]any
// to satisfy existing callers that assert concrete types.
func normalizeShared(v interface{}) interface{} {
	switch vv := v.(type) {
	case []interface{}:
		// Check if it's a slice of maps; convert to []map[string]interface{}
		converted := make([]map[string]interface{}, 0, len(vv))
		for _, it := range vv {
			if m, ok := it.(map[string]interface{}); ok {
				converted = append(converted, m)
			} else {
				// Not homogeneous; return original
				return v
			}
		}
		return converted
	default:
		return v
	}
}

// GetAllSharedData returns all shared memory data
func (t *Team) GetAllSharedData() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	result := make(map[string]interface{})
	for k, v := range t.sharedMemory {
		result[k] = v
	}
	return result
}
