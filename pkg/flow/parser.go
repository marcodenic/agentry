package flow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type File struct {
	Agents map[string]Agent `yaml:"agents"`
	Tasks  []Task           `yaml:"tasks"`
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
	if len(f.Agents) == 0 {
		return nil, errors.New("no agents defined")
	}
	for i, t := range f.Tasks {
		if err := validateTask(t, f.Agents); err != nil {
			return nil, fmt.Errorf("task %d: %w", i, err)
		}
	}
	return &f, nil
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
