package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func searchReplaceSpec() builtinSpec {
	return builtinSpec{
		Desc: "Search and replace text in a file with optional regex support",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to edit",
				},
				"search": map[string]any{
					"type":        "string",
					"description": "Text or regex pattern to search for",
				},
				"replace": map[string]any{
					"type":        "string",
					"description": "Replacement text (supports regex capture groups if using regex)",
				},
				"regex": map[string]any{
					"type":        "boolean",
					"description": "Whether to treat search as a regex pattern (default: false)",
					"default":     false,
				},
				"max_replacements": map[string]any{
					"type":        "integer",
					"description": "Maximum number of replacements to make (default: -1 for all)",
					"default":     -1,
				},
			},
			"required": []string{"path", "search", "replace"},
			"example": map[string]any{
				"path":    "src/main.go",
				"search":  "old_function",
				"replace": "new_function",
			},
		},
		Exec: searchReplaceExec,
	}
}

func registerSearchReplaceBuiltins(reg *builtinRegistry) {
	reg.add("search_replace", searchReplaceSpec())
}

func searchReplaceExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	search, _ := args["search"].(string)
	replace, _ := args["replace"].(string)
	isRegex, _ := args["regex"].(bool)
	maxReplacements, _ := getIntArg(args, "max_replacements", -1)

	path = absPath(path)
	if err := checkForOverwrite(path); err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)
	var newContent string
	var replacements int

	if isRegex {
		re, err := regexp.Compile(search)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}
		if maxReplacements == -1 {
			newContent = re.ReplaceAllString(originalContent, replace)
			replacements = len(re.FindAllString(originalContent, -1))
		} else {
			matches := re.FindAllStringIndex(originalContent, -1)
			if len(matches) > maxReplacements {
				matches = matches[:maxReplacements]
			}
			replacements = len(matches)
			newContent = originalContent
			for i := len(matches) - 1; i >= 0; i-- {
				match := matches[i]
				matchText := originalContent[match[0]:match[1]]
				replacement := re.ReplaceAllString(matchText, replace)
				newContent = newContent[:match[0]] + replacement + newContent[match[1]:]
			}
		}
	} else {
		if maxReplacements == -1 {
			newContent = strings.ReplaceAll(originalContent, search, replace)
			replacements = strings.Count(originalContent, search)
		} else {
			newContent = originalContent
			for i := 0; i < maxReplacements; i++ {
				if idx := strings.Index(newContent, search); idx != -1 {
					newContent = newContent[:idx] + replace + newContent[idx+len(search):]
					replacements++
				} else {
					break
				}
			}
		}
	}

	if replacements == 0 {
		return fmt.Sprintf(`{"path": "%s", "replacements": 0, "message": "No matches found"}`, path), nil
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	// Update viewed timestamp after modification so subsequent edits are allowed
	_ = recordView(path)

	resultInfo := map[string]any{
		"path":          path,
		"search":        search,
		"replace":       replace,
		"regex":         isRegex,
		"replacements":  replacements,
		"original_size": len(originalContent),
		"new_size":      len(newContent),
	}

	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}
