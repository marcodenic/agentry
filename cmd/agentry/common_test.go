package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
)

func TestApplyOverridesDisableTools(t *testing.T) {
	cfg := &config.File{
		Permissions: config.Permissions{Tools: []string{"echo", "write"}},
		Tools:       []config.ToolManifest{{Name: "echo"}, {Name: "write"}},
	}
	opts := &commonOpts{disableTools: true}

	applyOverrides(cfg, opts)

	if cfg.Permissions.Tools != nil {
		t.Fatalf("expected permissions to be cleared, got %v", cfg.Permissions.Tools)
	}
	if len(cfg.Tools) != 2 {
		t.Fatalf("expected registered tools to remain untouched, got %d", len(cfg.Tools))
	}
}

func TestApplyOverridesAllowList(t *testing.T) {
	cfg := &config.File{
		Permissions: config.Permissions{Tools: []string{"echo", "write"}},
		Tools:       []config.ToolManifest{{Name: "echo"}, {Name: "write"}},
	}
	opts := &commonOpts{allowTools: "echo"}

	applyOverrides(cfg, opts)

	if len(cfg.Tools) != 1 || cfg.Tools[0].Name != "echo" {
		t.Fatalf("expected only echo tool to remain, got %#v", cfg.Tools)
	}
	if len(cfg.Permissions.Tools) != 1 || cfg.Permissions.Tools[0] != "echo" {
		t.Fatalf("expected permissions to match allow list, got %v", cfg.Permissions.Tools)
	}
}

func TestApplyOverridesDenyList(t *testing.T) {
	cfg := &config.File{
		Permissions: config.Permissions{Tools: []string{"echo", "write"}},
		Tools:       []config.ToolManifest{{Name: "echo"}, {Name: "write"}},
	}
	opts := &commonOpts{denyTools: "write"}

	applyOverrides(cfg, opts)

	if len(cfg.Tools) != 1 || cfg.Tools[0].Name != "echo" {
		t.Fatalf("expected write tool to be removed, got %#v", cfg.Tools)
	}
	if len(cfg.Permissions.Tools) != 1 || cfg.Permissions.Tools[0] != "echo" {
		t.Fatalf("expected permissions to drop denied tool, got %v", cfg.Permissions.Tools)
	}
}

func TestParseCommonParsesFlags(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "custom.yaml")

	opts, args := parseCommon("agentry", []string{
		"--config", cfgPath,
		"--debug",
		"--disable-tools",
		"--allow-tools", "echo,write",
		"--deny-tools", "rm",
		"--max-iter", "5",
		"--http-timeout", "42",
		"prompt",
		"now",
	})

	if opts.configPath != cfgPath {
		t.Fatalf("expected config path %q, got %q", cfgPath, opts.configPath)
	}
	if !opts.debug {
		t.Fatalf("expected debug flag to be true")
	}
	if !opts.disableTools {
		t.Fatalf("expected disableTools flag to be true")
	}
	if opts.allowTools != "echo,write" {
		t.Fatalf("expected allowTools to match input, got %q", opts.allowTools)
	}
	if opts.denyTools != "rm" {
		t.Fatalf("expected denyTools to match input, got %q", opts.denyTools)
	}
	if opts.maxIter != 5 {
		t.Fatalf("expected maxIter 5, got %d", opts.maxIter)
	}
	if opts.httpTimeout != 42 {
		t.Fatalf("expected httpTimeout 42, got %d", opts.httpTimeout)
	}
	if len(args) != 2 || args[0] != "prompt" || args[1] != "now" {
		t.Fatalf("unexpected remaining args: %#v", args)
	}
}

func TestParseCommonTuiUsesFirstArgAsConfigPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	opts, args := parseCommon("tui", []string{"./config.yaml"})

	if opts.configPath != "./config.yaml" {
		t.Fatalf("expected config path to come from positional arg, got %q", opts.configPath)
	}
	if len(args) != 1 || args[0] != "./config.yaml" {
		t.Fatalf("expected positional args to be preserved, got %#v", args)
	}
}

func TestParseCommonDefaultsToDotAgentry(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	opts, args := parseCommon("agentry", nil)

	if opts.configPath != ".agentry.yaml" {
		t.Fatalf("expected default config path, got %q", opts.configPath)
	}
	if len(args) != 0 {
		t.Fatalf("expected no remaining args, got %#v", args)
	}
}
