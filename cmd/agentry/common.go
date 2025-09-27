package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/model"
)

type commonOpts struct {
	configPath   string
	debug        bool
	disableTools bool
	allowTools   string
	denyTools    string

	// New flags (prefer flags over env vars)
	maxIter     int // 0 = unlimited
	httpTimeout int // seconds
}

func parseCommon(name string, args []string) (*commonOpts, []string) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	opts := newCommonOpts()
	opts.bindFlags(fs)
	_ = fs.Parse(args)
	opts.configPath = resolveConfigPath(name, opts.configPath, fs)
	return opts, fs.Args()
}

func newCommonOpts() *commonOpts {
	return &commonOpts{}
}

func (o *commonOpts) bindFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.configPath, "config", "", "path to .agentry.yaml")
	fs.BoolVar(&o.debug, "debug", false, "enable debug output")
	fs.BoolVar(&o.disableTools, "disable-tools", false, "disable tool filtering entirely")
	fs.StringVar(&o.allowTools, "allow-tools", "", "comma-separated list of additional tools to include")
	fs.StringVar(&o.denyTools, "deny-tools", "", "comma-separated list of tools to exclude")
	fs.IntVar(&o.maxIter, "max-iter", 0, "limit agent iterations (0=unlimited)")
	fs.IntVar(&o.httpTimeout, "http-timeout", 300, "HTTP client timeout in seconds")
}

func resolveConfigPath(cmd string, explicit string, fs *flag.FlagSet) string {
	if explicit != "" {
		return explicit
	}
	if cmd == "tui" && fs.NArg() > 0 {
		return fs.Arg(0)
	}
	if path, ok := discoverConfigPath(".agentry.yaml"); ok {
		return path
	}
	return ".agentry.yaml"
}

func discoverConfigPath(candidate string) (string, bool) {
	if _, err := os.Stat(candidate); err == nil {
		return candidate, true
	}
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	if exeDir := filepath.Dir(exe); exeDir != "" {
		path := filepath.Join(exeDir, candidate)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func applyOverrides(cfg *config.File, o *commonOpts) {
	if o.debug {
		debug.EnableDebug()
	}

	if o.disableTools {
		cfg.Permissions.Tools = nil
	}
	if allow := parseCSV(o.allowTools); len(allow) > 0 {
		applyToolAllowList(cfg, allow)
	}
	if deny := parseCSV(o.denyTools); len(deny) > 0 {
		applyToolDenyList(cfg, deny)
	}

	if o.httpTimeout > 0 {
		model.SetHTTPTimeout(o.httpTimeout)
	}
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func applyToolAllowList(cfg *config.File, allow []string) {
	allowSet := make(map[string]struct{}, len(allow))
	for _, name := range allow {
		allowSet[name] = struct{}{}
	}
	filtered := make([]config.ToolManifest, 0, len(cfg.Tools))
	for _, tool := range cfg.Tools {
		if _, ok := allowSet[tool.Name]; ok {
			filtered = append(filtered, tool)
		}
	}
	cfg.Tools = filtered
	cfg.Permissions.Tools = allow
}

func applyToolDenyList(cfg *config.File, deny []string) {
	denySet := make(map[string]struct{}, len(deny))
	for _, name := range deny {
		denySet[name] = struct{}{}
	}
	filtered := make([]config.ToolManifest, 0, len(cfg.Tools))
	for _, tool := range cfg.Tools {
		if _, blocked := denySet[tool.Name]; !blocked {
			filtered = append(filtered, tool)
		}
	}
	cfg.Tools = filtered

	if cfg.Permissions.Tools == nil {
		return
	}
	perms := cfg.Permissions.Tools[:0]
	for _, name := range cfg.Permissions.Tools {
		if _, blocked := denySet[name]; !blocked {
			perms = append(perms, name)
		}
	}
	cfg.Permissions.Tools = perms
}
