package tests

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
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
	// generate signing key
	if out, err := exec.Command("gpg", "--batch", "--passphrase", "", "--quick-gen-key", "tester@example.com", "default", "default", "never").CombinedOutput(); err != nil {
		t.Fatalf("gen key: %v: %s", err, out)
	}
	dataPath := filepath.Join(dir, "data.txt")
	os.WriteFile(dataPath, []byte("p|"+pluginPath+"|"+hex.EncodeToString(sum[:])), 0644)
	sigBytes, err := exec.Command("gpg", "--batch", "--yes", "--local-user", "tester@example.com", "--output", "-", "--detach-sign", dataPath).Output()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	sig := base64.StdEncoding.EncodeToString(sigBytes)
	idx := plugin.Registry{Plugins: []plugin.RegistryEntry{{Name: "p", URL: pluginPath, SHA256: hex.EncodeToString(sum[:]), Sig: sig}}}
	idxPath := filepath.Join(dir, "index.json")
	b, _ := json.Marshal(idx)
	if err := os.WriteFile(idxPath, b, 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	pubPath := filepath.Join(dir, "pub.asc")
	if out, err := exec.Command("gpg", "--armor", "--output", pubPath, "--export", "tester@example.com").CombinedOutput(); err != nil {
		t.Fatalf("export: %v: %s", err, out)
	}

	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	os.Setenv("AGENTRY_REGISTRY_GPG_KEYRING", pubPath)
	defer os.Unsetenv("AGENTRY_REGISTRY_GPG_KEYRING")

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
