package env

import (
	"os"
	"path/filepath"
	"testing"
)

func unsetEnvForTest(t *testing.T, key string) {
	prev, had := os.LookupEnv(key)
	if had {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			_ = os.Setenv(key, prev)
		})
	} else {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("ensure unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			_ = os.Unsetenv(key)
		})
	}
}

func TestLoadFindsEnvInParentDirectory(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(parent, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	envFile := filepath.Join(parent, ".env.local")
	if err := os.WriteFile(envFile, []byte("FROM_PARENT=parent-value\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	if err := os.Chdir(child); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	unsetEnvForTest(t, "AGENTRY_ENV_FILE")
	unsetEnvForTest(t, "FROM_PARENT")

	Load()

	if got := os.Getenv("FROM_PARENT"); got != "parent-value" {
		t.Fatalf("expected env loaded from parent, got %q", got)
	}
}

func TestLoadHonorsEnvFileOverride(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "custom.env")
	if err := os.WriteFile(envPath, []byte("CUSTOM_ENV=value\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Setenv("AGENTRY_ENV_FILE", "custom.env")
	unsetEnvForTest(t, "CUSTOM_ENV")

	Load()

	if got := os.Getenv("CUSTOM_ENV"); got != "value" {
		t.Fatalf("expected env override file to load, got %q", got)
	}
}

func TestLoadWithExplicitFilename(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "explicit.env")
	if err := os.WriteFile(envPath, []byte("EXPLICIT=1\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	unsetEnvForTest(t, "EXPLICIT")

	Load("explicit.env")

	if got := os.Getenv("EXPLICIT"); got != "1" {
		t.Fatalf("expected explicit env file to load, got %q", got)
	}
}

func TestIntBoolFloatHelpers(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	if got := Int("TEST_INT", 0); got != 42 {
		t.Fatalf("expected int helper to parse value, got %d", got)
	}

	t.Setenv("TEST_INT", "not-a-number")
	if got := Int("TEST_INT", 7); got != 7 {
		t.Fatalf("expected default for invalid int, got %d", got)
	}

	t.Setenv("TEST_BOOL", "YES")
	if got := Bool("TEST_BOOL", false); !got {
		t.Fatalf("expected bool helper to understand YES")
	}
	t.Setenv("TEST_BOOL", "off")
	if got := Bool("TEST_BOOL", true); got {
		t.Fatalf("expected bool helper to understand off")
	}
	t.Setenv("TEST_BOOL", "unknown")
	if got := Bool("TEST_BOOL", true); !got {
		t.Fatalf("expected default bool when value unrecognized")
	}

	t.Setenv("TEST_FLOAT", "3.14")
	if got := Float("TEST_FLOAT", 0); got != 3.14 {
		t.Fatalf("expected float helper to parse value, got %f", got)
	}
	t.Setenv("TEST_FLOAT", "bad")
	if got := Float("TEST_FLOAT", 2.5); got != 2.5 {
		t.Fatalf("expected default for invalid float, got %f", got)
	}
}
