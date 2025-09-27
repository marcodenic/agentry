package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/model"
)

type commonOpts struct {
	configPath     string
	theme          string
	keybindsPath   string
	credsPath      string
	mcpFlag        string
	saveID         string
	resumeID       string
	ckptID         string
	port           string
	debug          bool
	disableTools   bool
	allowTools     string
	denyTools      string
	disableContext bool
	auditLog       string

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
	fs.StringVar(&o.theme, "theme", "", "theme name override")
	fs.StringVar(&o.keybindsPath, "keybinds", "", "path to keybinds json")
	fs.StringVar(&o.credsPath, "creds", "", "path to credentials json")
	fs.StringVar(&o.mcpFlag, "mcp", "", "comma-separated MCP servers")
	fs.StringVar(&o.saveID, "save-id", "", "save conversation state to this ID")
	fs.StringVar(&o.resumeID, "resume-id", "", "load conversation state from this ID")
	fs.StringVar(&o.ckptID, "checkpoint-id", "", "checkpoint session id")
	fs.StringVar(&o.port, "port", "", "HTTP server port")
	fs.BoolVar(&o.debug, "debug", false, "enable debug output")
	fs.BoolVar(&o.disableTools, "disable-tools", false, "disable tool filtering entirely")
	fs.StringVar(&o.allowTools, "allow-tools", "", "comma-separated list of additional tools to include")
	fs.StringVar(&o.denyTools, "deny-tools", "", "comma-separated list of tools to exclude")
	fs.BoolVar(&o.disableContext, "disable-context", false, "disable context pipeline")
	fs.StringVar(&o.auditLog, "audit-log", "", "path to audit log file")
	fs.IntVar(&o.maxIter, "max_iter", 0, "limit agent iterations (0=unlimited)")
	fs.IntVar(&o.maxIter, "max-iter", 0, "limit agent iterations (0=unlimited)")
	fs.IntVar(&o.httpTimeout, "http_timeout", 300, "HTTP client timeout in seconds")
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
	// Handle debug flag by enabling debug output dynamically
	if o.debug {
		debug.EnableDebug()
	}

	// Handle tool filtering flags by modifying config directly
	if o.disableTools {
		// Clear tool permissions to allow all tools
		cfg.Permissions.Tools = nil
	}
	if allow := parseCSV(o.allowTools); len(allow) > 0 {
		applyToolAllowList(cfg, allow)
	}
	if deny := parseCSV(o.denyTools); len(deny) > 0 {
		applyToolDenyList(cfg, deny)
	}

	// Handle context and audit flags
	if o.disableContext {
		os.Setenv("AGENTRY_DISABLE_CONTEXT", "1")
	}
	if o.auditLog != "" {
		os.Setenv("AGENTRY_AUDIT_LOG", o.auditLog)
	}

	if o.theme != "" {
		if cfg.Themes == nil {
			cfg.Themes = map[string]string{}
		}
		cfg.Themes["active"] = o.theme
		cfg.Theme = o.theme
	}
	if cfg.Theme != "" {
		os.Setenv("AGENTRY_THEME", cfg.Theme)
	}
	if o.keybindsPath != "" {
		if b, err := os.ReadFile(o.keybindsPath); err == nil {
			_ = json.Unmarshal(b, &cfg.Keybinds)
		}
	}
	if o.credsPath != "" {
		if b, err := os.ReadFile(o.credsPath); err == nil {
			_ = json.Unmarshal(b, &cfg.Credentials)
		}
	}
	if o.mcpFlag != "" {
		if cfg.MCPServers == nil {
			cfg.MCPServers = map[string]string{}
		}
		parts := parseCSV(o.mcpFlag)
		for i, p := range parts {
			cfg.MCPServers[fmt.Sprintf("srv%d", i+1)] = p
		}
	}

	// Apply model HTTP timeout (flags take precedence over env)
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
