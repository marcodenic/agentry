package tests

import (
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
)

func TestCollectorEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/conf.yaml"
	os.WriteFile(path, []byte("metrics: true\n"), 0644)
	os.Setenv("AGENTRY_COLLECTOR", "127.0.0.1:4318")
	defer os.Unsetenv("AGENTRY_COLLECTOR")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Collector != "127.0.0.1:4318" {
		t.Fatalf("collector not overridden: %s", cfg.Collector)
	}
}
