package prompt

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestSectionizeProducesStableEnvelope(t *testing.T) {
	reg := tool.Registry{
		"beta":  tool.New("beta", "", func(context.Context, map[string]any) (string, error) { return "", nil }),
		"alpha": tool.New("alpha", "Alpha description", func(context.Context, map[string]any) (string, error) { return "ok", nil }),
	}

	extras := map[string]string{
		"tool_guidance": "Use tools responsibly",
		"agents":        "coder",
		"output_format": "json",
	}

	out := Sectionize("Be helpful", reg, extras)

	const expected = `<agentry>
<prompt>
Be helpful
</prompt>
<agents>
coder
</agents>
<tools>
Use tools responsibly

- alpha: Alpha description
- beta
</tools>
<output_format>
json
</output_format>
</agentry>
`

	if out != expected {
		t.Fatalf("unexpected sectionized output:\n%s", out)
	}
}
