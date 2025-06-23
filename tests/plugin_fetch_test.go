package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/plugin"
)

func TestFetchVerifiesChecksum(t *testing.T) {
	dir := t.TempDir()
	data := []byte("hello")
	pluginPath := filepath.Join(dir, "p.txt")
	if err := os.WriteFile(pluginPath, data, 0644); err != nil {
		t.Fatalf("write plugin: %v", err)
	}
	sum := sha256.Sum256(data)
	idx := plugin.Registry{Plugins: []plugin.RegistryEntry{{Name: "p", URL: pluginPath, SHA256: hex.EncodeToString(sum[:])}}}
	idxPath := filepath.Join(dir, "index.json")
	b, _ := json.Marshal(idx)
	if err := os.WriteFile(idxPath, b, 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	out, err := plugin.Fetch(idxPath, "p")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("unexpected data %s", string(got))
	}
}
