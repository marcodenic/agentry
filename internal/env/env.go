package env

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Load loads variables from .env.local if the file exists.
func Load() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		path := filepath.Join(dir, ".env.local")
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}
