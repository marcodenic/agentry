package team

import (
	"os"

	"github.com/marcodenic/agentry/internal/debug"
)

func isTUI() bool   { return os.Getenv("AGENTRY_TUI_MODE") == "1" }
func isDebug() bool { return debug.DebugEnabled }

// logToFile records team events through the central debug logger.
func logToFile(message string) {
	debug.LogToFile("TEAM", "%s", message)
}

// debugPrintf uses the shared debug package for optional stderr output.
func debugPrintf(format string, v ...interface{}) {
	debug.Printf(format, v...)
}
