package tool

import "testing"

func TestFileOperationToolsInRegistry(t *testing.T) {
	registry := DefaultRegistry()
	
	expectedTools := []string{
		"read_lines",
		"edit_range", 
		"insert_at",
		"search_replace",
		"fileinfo",
		"view",
		"create",
	}
	
	for _, toolName := range expectedTools {
		tool, exists := registry.Use(toolName)
		if !exists {
			t.Errorf("Tool %s not found in registry", toolName)
			continue
		}
		
		if tool.Name() != toolName {
			t.Errorf("Tool name mismatch: expected %s, got %s", toolName, tool.Name())
		}
		
		schema := tool.JSONSchema()
		if schema == nil {
			t.Errorf("Tool %s has no schema", toolName)
		}
	}
}
