package tool

// TerminalAware is an optional interface a Tool can implement to signal that
// a successful execution should normally terminate the agent loop without an
// additional model reflection pass. Example: the "agent" delegation tool
// returns a final answer string from the delegated agent.
type TerminalAware interface {
	Tool
	Terminal() bool
}

// terminalTool wraps a Tool and marks it terminal.
type terminalTool struct{ Tool }

func (t terminalTool) Terminal() bool { return true }

// MarkTerminal wraps a Tool so the agent runtime can detect it and finalize
// immediately after successful execution (if all tool calls in a step are
// terminal and there are no errors).
func MarkTerminal(t Tool) Tool { return terminalTool{t} }
