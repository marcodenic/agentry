package config

import "os"

// RuntimeEnv captures the environment interactions required for
// applying runtime configuration overrides. It is intentionally small
// so tests can provide a deterministic implementation.
type RuntimeEnv interface {
	Getenv(key string) string
	Setenv(key, value string) error
	Unsetenv(key string) error
}

type osRuntimeEnv struct{}

func (osRuntimeEnv) Getenv(key string) string {
	return os.Getenv(key)
}

func (osRuntimeEnv) Setenv(key, value string) error {
	return os.Setenv(key, value)
}

func (osRuntimeEnv) Unsetenv(key string) error {
	return os.Unsetenv(key)
}

// OSEnv returns a RuntimeEnv backed by the real process environment.
func OSEnv() RuntimeEnv {
	return osRuntimeEnv{}
}

// RuntimeSettings describes the pieces of runtime configuration that
// require environment updates.
type RuntimeSettings struct {
	Theme          string
	AuditLogPath   string
	DisableContext bool
}

// RuntimeMutator applies RuntimeSettings to a given RuntimeEnv.
type RuntimeMutator struct {
	env RuntimeEnv
}

// NewRuntimeMutator creates a RuntimeMutator using the provided
// RuntimeEnv. A nil env falls back to the real OS environment.
func NewRuntimeMutator(env RuntimeEnv) *RuntimeMutator {
	if env == nil {
		env = osRuntimeEnv{}
	}
	return &RuntimeMutator{env: env}
}

// Env exposes the underlying RuntimeEnv, which is useful for callers
// that need to inspect existing values before applying overrides.
func (m *RuntimeMutator) Env() RuntimeEnv {
	return m.env
}

// Apply mutates the environment using the provided settings. Only
// non-empty values are written, and DisableContext=false removes the
// corresponding flag when present.
func (m *RuntimeMutator) Apply(settings RuntimeSettings) error {
	if settings.Theme != "" {
		if err := m.env.Setenv("AGENTRY_THEME", settings.Theme); err != nil {
			return err
		}
	}
	if settings.AuditLogPath != "" {
		if err := m.env.Setenv("AGENTRY_AUDIT_LOG", settings.AuditLogPath); err != nil {
			return err
		}
	}
	if settings.DisableContext {
		if err := m.env.Setenv("AGENTRY_DISABLE_CONTEXT", "1"); err != nil {
			return err
		}
	} else if current := m.env.Getenv("AGENTRY_DISABLE_CONTEXT"); current != "" {
		if err := m.env.Unsetenv("AGENTRY_DISABLE_CONTEXT"); err != nil {
			return err
		}
	}
	return nil
}
