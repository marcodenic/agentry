package tests

import "testing"

func TestSimpleMath(t *testing.T) {
	if 2+2 != 4 {
		t.Fatalf("expected 4, got something else")
	}
}
