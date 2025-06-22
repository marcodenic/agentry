package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/marcodenic/agentry/internal/plugin"
)

func TestInstallPluginUpdatesManifest(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	var cmds [][]string
	plugin.ExecCommand = func(name string, args ...string) *exec.Cmd {
		cmds = append(cmds, append([]string{name}, args...))
		return exec.Command("true")
	}
	defer func() { plugin.ExecCommand = exec.Command }()

	repo := "github.com/example/plugin"
	if err := plugin.Install(repo); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	b, err := os.ReadFile(plugin.ManifestFile)
	if err != nil {
		t.Fatalf("manifest read: %v", err)
	}
	var repos []string
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatalf("manifest json: %v", err)
	}
	if len(repos) != 1 || repos[0] != repo {
		t.Fatalf("unexpected manifest %v", repos)
	}

	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if cmds[0][0] != "git" || cmds[1][0] != "go" {
		t.Fatalf("unexpected commands %v", cmds)
	}
}
