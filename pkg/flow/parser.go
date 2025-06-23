package flow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type File struct {
	Presets []string         `yaml:"presets,omitempty"`
	Agents  map[string]Agent `yaml:"agents"`
	Tasks   []Task           `yaml:"tasks"`
}

type Agent struct {
	Model  string            `yaml:"model"`
	Prompt string            `yaml:"prompt,omitempty"`
	Tools  []string          `yaml:"tools,omitempty"`
	Env    map[string]string `yaml:"env,omitempty"`
}

type Task struct {
	Agent      string            `yaml:"agent,omitempty"`
	Input      string            `yaml:"input,omitempty"`
	Sequential []Task            `yaml:"sequential,omitempty"`
	Parallel   []Task            `yaml:"parallel,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
}

func merge(dst *File, src File) {
	if dst.Agents == nil {
		dst.Agents = map[string]Agent{}
	}
	for k, v := range src.Agents {
		dst.Agents[k] = v
	}
	dst.Tasks = append(dst.Tasks, src.Tasks...)
}

func resolvePreset(name, baseDir string) string {
	if filepath.IsAbs(name) {
		return name
	}
	if _, err := os.Stat(filepath.Join(baseDir, name)); err == nil {
		return filepath.Join(baseDir, name)
	}
	dir := baseDir
	for {
		p := filepath.Join(dir, "templates", name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "" {
			break
		}
		dir = parent
	}
	return filepath.Join(baseDir, name)
}

// Load reads and validates a flow file.
func Load(path string) (*File, error) {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		path = filepath.Join(path, ".agentry.flow.yaml")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(path)
	var out File
	for _, p := range f.Presets {
		pf, err := Load(resolvePreset(p, baseDir))
		if err != nil {
			return nil, err
		}
		merge(&out, *pf)
	}
	f.Presets = nil
	merge(&out, f)

	if len(out.Agents) == 0 {
		return nil, errors.New("no agents defined")
	}
	for i, t := range out.Tasks {
		if err := validateTask(t, out.Agents); err != nil {
			return nil, fmt.Errorf("task %d: %w", i, err)
		}
	}
	return &out, nil
}

func validateTask(t Task, agents map[string]Agent) error {
	if t.Agent == "" && len(t.Sequential) == 0 && len(t.Parallel) == 0 {
		return errors.New("task must define agent or subtasks")
	}
	if t.Agent != "" {
		if _, ok := agents[t.Agent]; !ok {
			return fmt.Errorf("undefined agent %q", t.Agent)
		}
	}
	for i, st := range t.Sequential {
		if err := validateTask(st, agents); err != nil {
			return fmt.Errorf("sequential[%d]: %w", i, err)
		}
	}
	for i, st := range t.Parallel {
		if err := validateTask(st, agents); err != nil {
			return fmt.Errorf("parallel[%d]: %w", i, err)
		}
	}
	return nil
}
