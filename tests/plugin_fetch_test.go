package tests

import (
	"crypto/ed25519"
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
	priv := ed25519.NewKeyFromSeed(make([]byte, 32))
	plugin.SignRegistry(&idx, priv)
	idxPath := filepath.Join(dir, "index.json")
	b, _ := json.Marshal(idx)
	if err := os.WriteFile(idxPath, b, 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	pubPath := filepath.Join(dir, "pub")
	os.WriteFile(pubPath, []byte(hex.EncodeToString(priv.Public().(ed25519.PublicKey))), 0644)

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	os.Setenv("AGENTRY_REGISTRY_PUBKEY", pubPath)
	defer os.Unsetenv("AGENTRY_REGISTRY_PUBKEY")

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
