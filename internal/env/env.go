package env

import (
	"log"
	"os"
	"path/filepath"

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
