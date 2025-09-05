package context

import "github.com/marcodenic/agentry/internal/model"

// Assembler orchestrates provider and budget components.
type Assembler struct {
	Provider Provider
	Budget   Budget
}

// Assemble builds final messages from input through provider and budget.
func (a Assembler) Assemble(input string) []model.ChatMessage {
	msgs := a.Provider.Provide(input)
	return a.Budget.Apply(msgs)
}

// AssembleWithTools builds messages and applies budgeting with tool schema overhead.
func (a Assembler) AssembleWithTools(input string, specs []model.ToolSpec) []model.ChatMessage {
	msgs := a.Provider.Provide(input)
	return a.Budget.ApplyWithTools(msgs, specs)
}
