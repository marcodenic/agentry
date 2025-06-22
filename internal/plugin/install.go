package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// ExecCommand allows tests to override exec.Command.
var ExecCommand = exec.Command

// ManifestFile stores installed plugin repositories.
const ManifestFile = ".agentry.plugins.json"

// Install clones a repository and installs it via go install.
func Install(repo string) error {
	dir, err := os.MkdirTemp("", "agentry-plugin-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if out, err := ExecCommand("git", "clone", repo, dir).CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %v: %s", err, out)
	}

	cmd := ExecCommand("go", "install")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go install: %v: %s", err, out)
	}

	var plugins []string
	if b, err := os.ReadFile(ManifestFile); err == nil {
		_ = json.Unmarshal(b, &plugins)
	}
	for _, p := range plugins {
		if p == repo {
			return nil
		}
	}
	plugins = append(plugins, repo)
	b, _ := json.MarshalIndent(plugins, "", "  ")
	return os.WriteFile(ManifestFile, b, 0644)
}
