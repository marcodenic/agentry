package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Load loads variables from .env.local if the file exists.
// Load searches for filename (default ".env.local") upward and loads variables.
func Load(filename ...string) {
	name := ".env.local"
	if len(filename) > 0 && filename[0] != "" {
		name = filename[0]
	} else if env := os.Getenv("AGENTRY_ENV_FILE"); env != "" {
		name = env
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("env load: %v", err)
		return
	}
	for {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("env load: %v", err)
			}
			break
		} else if !os.IsNotExist(err) {
			log.Printf("env stat: %v", err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

// Int retrieves an int from env or returns def if unset/invalid.
func Int(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

// Bool retrieves a boolean (accepts 1/0 true/false yes/no on/off) or def.
func Bool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return def
}

// Float retrieves a float64 or def.
func Float(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// WarnDeprecatedEnvVars checks for deprecated AGENTRY_* environment variables
// and prints deprecation warnings to stderr. Returns a slice of warned variable names.
func WarnDeprecatedEnvVars() []string {
	var warned []string

	for _, env := range os.Environ() {
		// Split on first '=' to get key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]

		// Check if it starts with AGENTRY_ but is not exactly AGENTRY_CONFIG
		if strings.HasPrefix(key, "AGENTRY_") && key != "AGENTRY_CONFIG" {
			fmt.Fprintf(os.Stderr, "Deprecation: environment variable %s is deprecated and ignored. Use YAML config; AGENTRY_CONFIG is the only supported env var.\n", key)
			warned = append(warned, key)
		}
	}

	return warned
}
