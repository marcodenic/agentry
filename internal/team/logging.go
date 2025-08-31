package team

import (
    "fmt"
    "log"
    "os"

    "github.com/marcodenic/agentry/internal/debug"
)

// logToFile logs the message to a file (only if explicitly enabled and not in TUI mode)
func logToFile(message string) {
    if os.Getenv("AGENTRY_TUI_MODE") == "1" {
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
    if (os.Getenv("AGENTRY_DEBUG") == "1" || os.Getenv("AGENTRY_DEBUG") == "true") && os.Getenv("AGENTRY_TUI_MODE") != "1" {
        fmt.Fprintf(os.Stderr, format, v...)
    }
}
