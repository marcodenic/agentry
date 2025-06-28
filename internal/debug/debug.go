package debug

import (
	"io"
	"log"
	"os"
)

var (
	// DebugEnabled controls whether debug output is enabled
	DebugEnabled bool
	// DebugLogger is the logger used for debug output
	DebugLogger *log.Logger
)

func init() {
	// Check environment variable for debug mode
	DebugEnabled = os.Getenv("AGENTRY_DEBUG") == "1" || os.Getenv("AGENTRY_DEBUG") == "true"
	
	if DebugEnabled {
		// In debug mode, output to stderr
		DebugLogger = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
	} else {
		// In non-debug mode, discard output
		DebugLogger = log.New(io.Discard, "", 0)
	}
}

// Printf writes debug output if debug mode is enabled
func Printf(format string, v ...interface{}) {
	if DebugEnabled {
		DebugLogger.Printf(format, v...)
	}
}

// SetTUIMode disables debug output to avoid interfering with TUI
func SetTUIMode(enabled bool) {
	if enabled && DebugEnabled {
		// In TUI mode, redirect debug output to a file or discard it
		DebugLogger = log.New(io.Discard, "", 0)
	} else if DebugEnabled {
		// Restore normal debug output
		DebugLogger = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
	}
}
