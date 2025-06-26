package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/pkg/sbox"
)

var osType = runtime.GOOS
var shellCmd, shellFlag string

func init() {
	if osType == "windows" {
		shellCmd = "powershell.exe"
		shellFlag = "-Command"
	} else {
		shellCmd = "bash"
		shellFlag = "-c"
	}
}

func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	wd, _ := os.Getwd()
	return filepath.Join(wd, p)
}

var ErrUnknownManifest = errors.New("unknown tool manifest")
var ErrUnknownBuiltin = errors.New("unknown builtin tool")
var ErrToolDenied = errors.New("tool not permitted")

var allowedTools map[string]struct{}

// SetPermissions configures which tools may execute. Nil or empty slice allows all.
func SetPermissions(list []string) {
	if len(list) == 0 {
		allowedTools = nil
		return
	}
	allowedTools = make(map[string]struct{}, len(list))
	for _, n := range list {
		allowedTools[n] = struct{}{}
	}
}

func permitted(name string) bool {
	if allowedTools == nil {
		return true
	}
	_, ok := allowedTools[name]
	return ok
}

// viewedFiles tracks file paths read via the view builtin along with their
// modification time. It is used to prevent overwriting files that have changed
// on disk since they were last viewed.
var viewedFiles sync.Map

// confirmOverwrite toggles interactive confirmation before overwriting. Set the
// AGENTRY_CONFIRM environment variable to any non-empty value to enable.
var confirmOverwrite = os.Getenv("AGENTRY_CONFIRM") != ""

func recordView(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	viewedFiles.Store(path, info.ModTime())
	return nil
}

func confirm(msg string) bool {
	if !confirmOverwrite {
		return false
	}
	fmt.Printf("%s [y/N]: ", msg)
	rd := bufio.NewReader(os.Stdin)
	line, _ := rd.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

func checkForOverwrite(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	v, ok := viewedFiles.Load(path)
	if !ok {
		if confirm("overwrite " + path + " without viewing?") {
			return nil
		}
		return fmt.Errorf("file %s must be viewed before modification", path)
	}
	if mod := v.(time.Time); !info.ModTime().Equal(mod) {
		if confirm("file " + path + " changed since viewed, overwrite?") {
			return nil
		}
		return fmt.Errorf("file %s changed since viewed", path)
	}
	return nil
}

// IsBuiltinTool checks if the given name is a builtin tool
func IsBuiltinTool(name string) bool {
	_, exists := builtinMap[name]
	return exists
}

// DefaultRegistry returns all builtin tools.
func DefaultRegistry() Registry {
	r := make(Registry, len(builtinMap))
	for n, s := range builtinMap {
		r[n] = NewWithSchema(n, s.Desc, s.Schema, s.Exec)
	}
	return r
}

func FromManifest(m config.ToolManifest) (Tool, error) {
	// ensure only one of builtin, http or command is specified
	count := 0
	if m.Type != "" {
		count++
	}
	if m.HTTP != "" {
		count++
	}
	if m.Command != "" {
		count++
	}
	if count != 1 {
		return nil, ErrUnknownManifest
	}

	// Builtin Go tools
	if m.Type == "builtin" {
		spec, ok := builtinMap[m.Name]
		if !ok {
			return nil, ErrUnknownBuiltin
		}
		desc := m.Description
		if desc == "" {
			desc = spec.Desc
		}
		tl := NewWithSchema(m.Name, desc, spec.Schema, spec.Exec)
		if st, ok := tl.(*simpleTool); ok {
			allowed := true
			if m.Permissions.Allow != nil {
				allowed = *m.Permissions.Allow
			}
			st.allowed = allowed
		}
		return tl, nil
	}

	// HTTP tools
	if m.HTTP != "" {
		tl := NewWithSchema(m.Name, m.Description, map[string]any{"type": "object", "properties": map[string]any{}}, func(ctx context.Context, args map[string]any) (string, error) {
			b, err := json.Marshal(args)
			if err != nil {
				return "", err
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.HTTP, bytes.NewReader(b))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			rb, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			return string(rb), nil
		})
		if st, ok := tl.(*simpleTool); ok {
			allowed := true
			if m.Permissions.Allow != nil {
				allowed = *m.Permissions.Allow
			}
			st.allowed = allowed
		}
		return tl, nil
	}

	// Shell command tools
	if m.Command != "" {
		tl := NewWithSchema(m.Name, m.Description, map[string]any{"type": "object", "properties": map[string]any{}}, func(ctx context.Context, args map[string]any) (string, error) {
			if !m.Privileged {
				return ExecSandbox(ctx, m.Command, sbox.Options{
					Engine:   m.Engine,
					Net:      m.Net,
					CPULimit: m.CPULimit,
					MemLimit: m.MemLimit,
				})
			}
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(ctx, "cmd", "/C", m.Command)
			} else {
				cmd = exec.CommandContext(ctx, "sh", "-c", m.Command)
			}
			out, err := cmd.CombinedOutput()
			return string(out), err
		})
		if st, ok := tl.(*simpleTool); ok {
			allowed := true
			if m.Permissions.Allow != nil {
				allowed = *m.Permissions.Allow
			}
			st.allowed = allowed
		}
		return tl, nil
	}

	return nil, ErrUnknownManifest
}

func parsePatchFiles(patchStr string) []string {
	var files []string
	for _, line := range strings.Split(patchStr, "\n") {
		if strings.HasPrefix(line, "+++ ") {
			f := strings.TrimPrefix(line, "+++ ")
			f = strings.TrimPrefix(f, "b/")
			if f != "/dev/null" && f != "" {
				files = append(files, f)
			}
		}
	}
	return files
}
