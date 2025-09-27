package glyphs

import (
	"strings"
	"testing"
)

func TestStatusConstants(t *testing.T) {
	if StatusPending.Name != "pending" || StatusPending.Glyph == "" {
		t.Fatalf("expected pending status to be defined")
	}
	if StatusError.Glyph != Crossmark {
		t.Fatalf("expected status error glyph to equal crossmark")
	}
}

func TestStyledGlyphsContainBaseRune(t *testing.T) {
	if out := GreenCheckmark(); !strings.Contains(out, Checkmark) {
		t.Fatalf("expected green checkmark to contain base glyph")
	}
	if out := RedCrossmark(); !strings.Contains(out, Crossmark) {
		t.Fatalf("expected red crossmark to contain base glyph")
	}
	if out := OrangeLightning(); !strings.Contains(out, Lightning) {
		t.Fatalf("expected orange lightning to contain base glyph")
	}
}
