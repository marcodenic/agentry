package tests

import (
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
)

func TestLoadPermissions(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/conf.yaml"
	os.WriteFile(path, []byte("permissions:\n  tools:\n    - echo\n    - ls\n"), 0644)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cfg.Permissions.Tools) != 2 {
		t.Fatalf("unexpected permissions: %#v", cfg.Permissions.Tools)
	}
}
