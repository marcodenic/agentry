package team

import (
	"fmt"
	"log"
	"os"

	"github.com/marcodenic/agentry/internal/debug"
)

func isTUI() bool { return os.Getenv("AGENTRY_TUI_MODE") == "1" }
func isDebug() bool {
	d := os.Getenv("AGENTRY_DEBUG")
	return d == "1" || d == "true"
}

// logToFile logs the message to a file (always log important team events)
func logToFile(message string) {
	// Always log team communications to the debug system
	debug.LogToFile("TEAM", "%s", message)
	
	// Legacy file logging for backward compatibility
	if isTUI() {
		return
	}
	if !debug.IsCommLogEnabled() {
		return
	}
	file, err := os.OpenFile("agent_communication.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	log.New(file, "", log.LstdFlags).Println(message)
}

// debugPrintf prints debug information only when debug is enabled and not in TUI mode
func debugPrintf(format string, v ...interface{}) {
	if isDebug() && !isTUI() {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}
