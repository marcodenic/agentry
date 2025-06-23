package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
