package plugin

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Registry lists available plugins.
type Registry struct {
	Plugins []RegistryEntry `json:"plugins"`
}

// RegistryEntry describes a plugin artifact.
type RegistryEntry struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Sig    string `json:"sig,omitempty"`
}

// Fetch downloads a plugin by name from indexPath, verifies the SHA256 sum,
// and writes the file to the current directory. It returns the saved filename.
func Fetch(indexPath, name string) (string, error) {
	data, err := readBytes(indexPath)
	if err != nil {
		return "", err
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return "", err
	}
	var ent *RegistryEntry
	for i := range reg.Plugins {
		if reg.Plugins[i].Name == name {
			ent = &reg.Plugins[i]
			break
		}
	}
	if ent == nil {
		return "", fmt.Errorf("plugin %s not found", name)
	}
	if ring := os.Getenv("AGENTRY_REGISTRY_GPG_KEYRING"); ring != "" {
		if err := verifyGPG(*ent, ring); err != nil {
			return "", err
		}
	} else if pubPath := os.Getenv("AGENTRY_REGISTRY_PUBKEY"); pubPath != "" {
		pubData, err := os.ReadFile(pubPath)
		if err != nil {
			return "", err
		}
		pub, err := hex.DecodeString(strings.TrimSpace(string(pubData)))
		if err != nil {
			return "", err
		}
		if !VerifySignature(*ent, ed25519.PublicKey(pub)) {
			return "", fmt.Errorf("signature mismatch for %s", name)
		}
	}
	b, err := readBytes(ent.URL)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != ent.SHA256 {
		return "", fmt.Errorf("sha256 mismatch for %s", name)
	}
	out := filepath.Base(ent.URL)
	if err := os.WriteFile(out, b, 0644); err != nil {
		return "", err
	}
	return out, nil
}

func readBytes(path string) ([]byte, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(path)
}

// SignRegistry signs all entries in reg using priv and updates the Sig field.
func SignRegistry(reg *Registry, priv ed25519.PrivateKey) {
	for i := range reg.Plugins {
		data := []byte(reg.Plugins[i].Name + "|" + reg.Plugins[i].URL + "|" + reg.Plugins[i].SHA256)
		sig := ed25519.Sign(priv, data)
		reg.Plugins[i].Sig = hex.EncodeToString(sig)
	}
}

// VerifySignature checks the entry signature using pub.
func VerifySignature(e RegistryEntry, pub ed25519.PublicKey) bool {
	b, err := hex.DecodeString(e.Sig)
	if err != nil || len(b) == 0 {
		return false
	}
	data := []byte(e.Name + "|" + e.URL + "|" + e.SHA256)
	return ed25519.Verify(pub, data, b)
}

// verifyGPG checks the signature using a GPG keyring path.
func verifyGPG(e RegistryEntry, keyring string) error {
	if e.Sig == "" {
		return fmt.Errorf("missing signature")
	}
	sig, err := base64.StdEncoding.DecodeString(e.Sig)
	if err != nil {
		return err
	}
	sigFile, err := os.CreateTemp("", "sig")
	if err != nil {
		return err
	}
	defer os.Remove(sigFile.Name())
	if _, err := sigFile.Write(sig); err != nil {
		return err
	}
	sigFile.Close()
	data := []byte(e.Name + "|" + e.URL + "|" + e.SHA256)
	dataFile, err := os.CreateTemp("", "data")
	if err != nil {
		return err
	}
	defer os.Remove(dataFile.Name())
	if _, err := dataFile.Write(data); err != nil {
		return err
	}
	dataFile.Close()
	cmd := exec.Command("gpg", "--batch", "--no-default-keyring", "--keyring", keyring, "--verify", sigFile.Name(), dataFile.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gpg verify: %v: %s", err, out)
	}
	return nil
}
