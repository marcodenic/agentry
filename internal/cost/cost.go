package cost

import "sync"

const CostPerToken = 0.000002

type Manager struct {
	mu            sync.Mutex
	ModelTokens   map[string]int
	ToolTokens    map[string]int
	BudgetTokens  int
	BudgetDollars float64
}

func New(budgetTokens int, budgetDollars float64) *Manager {
	return &Manager{ModelTokens: map[string]int{}, ToolTokens: map[string]int{}, BudgetTokens: budgetTokens, BudgetDollars: budgetDollars}
}

func (m *Manager) AddModel(name string, tokens int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ModelTokens[name] += tokens
	return m.overBudgetLocked()
}

func (m *Manager) AddTool(name string, tokens int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ToolTokens[name] += tokens
	return m.overBudgetLocked()
}

func (m *Manager) TotalTokens() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalTokensLocked()
}

func (m *Manager) totalTokensLocked() int {
	total := 0
	for _, v := range m.ModelTokens {
		total += v
	}
	for _, v := range m.ToolTokens {
		total += v
	}
	return total
}

func (m *Manager) TotalCost() float64 {
	return float64(m.TotalTokens()) * CostPerToken
}

func (m *Manager) OverBudget() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.overBudgetLocked()
}

func (m *Manager) overBudgetLocked() bool {
	if m.BudgetTokens > 0 && m.totalTokensLocked() > m.BudgetTokens {
		return true
	}
	if m.BudgetDollars > 0 && m.TotalCost() > m.BudgetDollars {
		return true
	}
	return false
}
