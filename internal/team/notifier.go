package team

import (
	"fmt"
	"os"
)

// delegationNotifier centralizes stderr and file logging so callers don't repeat guards.
type delegationNotifier struct{}

func (delegationNotifier) User(format string, args ...interface{}) {
	if isTUI() {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func (delegationNotifier) File(format string, args ...interface{}) {
	logToFile(fmt.Sprintf(format, args...))
}
