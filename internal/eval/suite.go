package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
)

type Case struct {
	Input    string `json:"input"`
	Contains string `json:"contains"`
}

type Suite struct {
	Cases []Case `json:"cases"`
}

func Run(t *testing.T, ag *core.Agent, path string) {
	b, _ := os.ReadFile(path)
	var s Suite
	_ = json.Unmarshal(b, &s)
	for _, c := range s.Cases {
		out, err := ag.Run(context.Background(), c.Input)
		passed := err == nil && strings.Contains(out, c.Contains)
		if t != nil {
			if !passed {
				t.Errorf("fail: %#v -> %q (%v)", c, out, err)
			}
		} else {
			fmt.Printf("%s -> %q\n", c.Input, out)
			if !passed {
				fmt.Printf("fail: %#v -> %q (%v)\n", c, out, err)
			}
		}
	}
}
