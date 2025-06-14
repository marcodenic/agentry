package env

import (
	"os"

	"github.com/joho/godotenv"
)

// Load loads variables from .env.local if the file exists.
func Load() {
	if _, err := os.Stat(".env.local"); err == nil {
		_ = godotenv.Load(".env.local")
	}
}
