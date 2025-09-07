package prompt

import (
    "sort"
    "strings"

    "github.com/marcodenic/agentry/internal/tool"
)

// Sectionize builds a very simple tagged system prompt envelope.
// It is intentionally lightweight: no XML parsing, no escaping.
// Order is stable so tests can assert structure.
func Sectionize(basePrompt string, reg tool.Registry, extras map[string]string) string {
    var body strings.Builder

    // <prompt>
    body.WriteString("<prompt>\n")
    body.WriteString(strings.TrimSpace(basePrompt))
    body.WriteString("\n</prompt>\n")

    // <agents> (placed before tools for visibility)
    if agents, ok := extras["agents"]; ok {
        body.WriteString("<agents>\n")
        body.WriteString(strings.TrimSpace(agents))
        body.WriteString("\n</agents>\n")
        // Prevent duplicate rendering in extras loop below
        delete(extras, "agents")
    } else {
        // Still render the tag to keep structure consistent
        body.WriteString("<agents>\n</agents>\n")
    }

    // <tools>
    body.WriteString("<tools>\n")
    if g, ok := extras["tool_guidance"]; ok {
        g = strings.TrimSpace(g)
        if g != "" {
            body.WriteString(g)
            body.WriteString("\n\n")
        }
        delete(extras, "tool_guidance")
    }
    names := make([]string, 0, len(reg))
    for name := range reg {
        names = append(names, name)
    }
    sort.Strings(names)
    for _, n := range names {
        t, ok := reg[n]
        if !ok {
            continue
        }
        desc := strings.TrimSpace(t.Description())
        if desc != "" {
            body.WriteString("- ")
            body.WriteString(n)
            body.WriteString(": ")
            body.WriteString(desc)
            body.WriteString("\n")
        } else {
            body.WriteString("- ")
            body.WriteString(n)
            body.WriteString("\n")
        }
    }
    body.WriteString("</tools>\n")

    // Optional extras as simple <key>value</key> blocks
    // (e.g., agents, output-format). Order by key for stability.
    if len(extras) > 0 {
        keys := make([]string, 0, len(extras))
        for k := range extras {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        for _, k := range keys {
            body.WriteString("<")
            body.WriteString(k)
            body.WriteString(">\n")
            body.WriteString(strings.TrimSpace(extras[k]))
            body.WriteString("\n</")
            body.WriteString(k)
            body.WriteString(">\n")
        }
    }

    // Wrap with a single top-level tag for clarity
    var out strings.Builder
    out.WriteString("<agentry>\n")
    out.WriteString(body.String())
    out.WriteString("</agentry>\n")
    return out.String()
}
