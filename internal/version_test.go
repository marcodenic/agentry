package agentry

import (
	"strings"
	"testing"
)

func TestVersionIsSemver(t *testing.T) {
	if Version == "" {
		t.Fatal("version must not be empty")
	}
	for i, part := range []byte(Version) {
		switch {
		case part >= '0' && part <= '9':
		case part == '.':
		default:
			t.Fatalf("version %q contains invalid character at %d", Version, i)
		}
	}
	if dotCount := strings.Count(Version, "."); dotCount != 2 {
		t.Fatalf("expected semantic version with two dots, got %q", Version)
	}
}
