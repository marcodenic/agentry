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
	// debugLevel stores the current debug level
	debugLevel string
)

func init() {
	// Check for consolidated debug level first
	debugLevel = os.Getenv("AGENTRY_DEBUG_LEVEL")

	// Backward compatibility with old flags
	if debugLevel == "" {
		if os.Getenv("AGENTRY_DEBUG") == "1" || os.Getenv("AGENTRY_DEBUG") == "true" {
			debugLevel = "debug"
		} else if os.Getenv("AGENTRY_COMM_LOG") == "1" {
			debugLevel = "trace" // communication logging is trace level
		} else if os.Getenv("AGENTRY_DEBUG_CONTEXT") == "1" {
			debugLevel = "debug"
		}
	}

	// Set debug enabled based on level
	DebugEnabled = debugLevel == "debug" || debugLevel == "trace"

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

// IsTraceEnabled returns true if trace-level debugging is enabled
func IsTraceEnabled() bool {
	return debugLevel == "trace"
}

// IsContextDebugEnabled returns true if context debugging is enabled
func IsContextDebugEnabled() bool {
	return debugLevel == "debug" || debugLevel == "trace"
}

// IsCommLogEnabled returns true if communication logging is enabled
func IsCommLogEnabled() bool {
	return debugLevel == "trace" || os.Getenv("AGENTRY_COMM_LOG") == "1"
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
