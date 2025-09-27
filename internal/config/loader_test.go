package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func TestLoadLayerPrecedence(t *testing.T) {
	root := t.TempDir()

	globalDir := filepath.Join(root, "global")
	writeTestFile(t, filepath.Join(globalDir, "config.yaml"), `memory: global-mem
permissions:
  tools: [global]
tools:
  - name: global
`)
	t.Setenv("AGENTRY_CONFIG_HOME", globalDir)

	projectDir := filepath.Join(root, "project")
	writeTestFile(t, filepath.Join(projectDir, "agentry.yaml"), `store: bolt://project
sandbox:
  engine: bubblewrap
tools:
  - name: project
`)

	mainPath := filepath.Join(projectDir, "agent.yaml")
	writeTestFile(t, mainPath, `tools:
  - name: main
budget:
  tokens: 10
include: [roles/base.yaml]
vector_store:
  type: qdrant
  url: http://example
`)

	cfg, err := Load(mainPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Memory != "global-mem" {
		t.Fatalf("expected memory from global config, got %q", cfg.Memory)
	}
	if cfg.Store != "bolt://project" {
		t.Fatalf("expected store from project config, got %q", cfg.Store)
	}
	if cfg.Sandbox.Engine != "bubblewrap" {
		t.Fatalf("expected sandbox engine from project config, got %q", cfg.Sandbox.Engine)
	}
	if len(cfg.Tools) != 1 || cfg.Tools[0].Name != "main" {
		t.Fatalf("expected main config tools to win, got %#v", cfg.Tools)
	}
	if cfg.Permissions.Tools == nil || len(cfg.Permissions.Tools) != 1 || cfg.Permissions.Tools[0] != "global" {
		t.Fatalf("expected permissions from global config, got %#v", cfg.Permissions.Tools)
	}
	if cfg.Budget.Tokens != 10 {
		t.Fatalf("expected budget tokens from main config, got %d", cfg.Budget.Tokens)
	}
	if len(cfg.Include) != 1 || cfg.Include[0] != "roles/base.yaml" {
		t.Fatalf("expected include list from main config, got %#v", cfg.Include)
	}
	if cfg.Vector.Type != "qdrant" || cfg.Vector.URL != "http://example" {
		t.Fatalf("unexpected vector store: %#v", cfg.Vector)
	}
}

func TestLoadIgnoresUnreadableGlobalConfig(t *testing.T) {
	root := t.TempDir()

	globalDir := filepath.Join(root, "global")
	globalPath := filepath.Join(globalDir, "config.yaml")
	writeTestFile(t, globalPath, "memory: should-not-win\n")
	if err := os.Chmod(globalPath, 0); err != nil {
		t.Skipf("chmod unsupported: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(globalPath, 0o644)
	})
	t.Setenv("AGENTRY_CONFIG_HOME", globalDir)

	projectDir := filepath.Join(root, "project")
	writeTestFile(t, filepath.Join(projectDir, "agentry.yaml"), "store: bolt://project\n")

	mainPath := filepath.Join(projectDir, "agent.yaml")
	writeTestFile(t, mainPath, "memory: main\n")

	cfg, err := Load(mainPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Memory != "main" {
		t.Fatalf("expected main config to win despite unreadable global config, got %q", cfg.Memory)
	}
	if cfg.Store != "bolt://project" {
		t.Fatalf("expected project config still applied, got %q", cfg.Store)
	}
}

func TestLoadMissingFileReturnsError(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatalf("expected error when loading missing file")
	}
}

func TestLoadInvalidYAMLReturnsError(t *testing.T) {
	mainPath := filepath.Join(t.TempDir(), "bad.yaml")
	writeTestFile(t, mainPath, "::: not yaml :::")

	if _, err := Load(mainPath); err == nil {
		t.Fatalf("expected error for invalid YAML input")
	}
}
