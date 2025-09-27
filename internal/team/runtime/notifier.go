package runtime

import (
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/debug"
)

type Notifier interface {
	User(format string, args ...interface{})
	File(format string, args ...interface{})
}

type TeamNotifier struct{}

func NewNotifier() TeamNotifier { return TeamNotifier{} }

func (TeamNotifier) User(format string, args ...interface{}) {
	if IsTUI() {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func (TeamNotifier) File(format string, args ...interface{}) {
	LogToFile(fmt.Sprintf(format, args...))
}

func IsTUI() bool { return os.Getenv("AGENTRY_TUI_MODE") == "1" }

func LogToFile(message string) {
	debug.LogToFile("TEAM", "%s", message)
}

func DebugPrintf(format string, v ...interface{}) {
	debug.Printf(format, v...)
}

func IsDebug() bool { return debug.DebugEnabled }
