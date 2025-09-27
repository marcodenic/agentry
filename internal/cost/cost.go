package cost

import "sync"

// TokenUsage represents token usage for a model call
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

type Manager struct {
	mu            sync.Mutex
	ModelUsage    map[string]TokenUsage // Model name -> token usage with input/output breakdown
	BudgetTokens  int
	BudgetDollars float64
	pricing       *PricingTable
}

func New(budgetTokens int, budgetDollars float64) *Manager {
	return &Manager{
		ModelUsage:    map[string]TokenUsage{},
		BudgetTokens:  budgetTokens,
		BudgetDollars: budgetDollars,
		pricing:       NewPricingTable(),
	}
}

// AddModelUsage adds token usage for a specific model with input/output breakdown
func (m *Manager) AddModelUsage(modelName string, inputTokens, outputTokens int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.ModelUsage[modelName]
	current.InputTokens += inputTokens
	current.OutputTokens += outputTokens
	m.ModelUsage[modelName] = current

	return m.overBudgetLocked()
}

func (m *Manager) TotalTokens() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalTokensLocked()
}

func (m *Manager) totalTokensLocked() int {
	total := 0
	for _, usage := range m.ModelUsage {
		total += usage.InputTokens + usage.OutputTokens
	}
	return total
}

func (m *Manager) totalInputTokensLocked() int {
	total := 0
	for _, usage := range m.ModelUsage {
		total += usage.InputTokens
	}
	return total
}

func (m *Manager) totalOutputTokensLocked() int {
	total := 0
	for _, usage := range m.ModelUsage {
		total += usage.OutputTokens
	}
	return total
}

// TotalInputTokens returns the total number of input tokens consumed across models.
func (m *Manager) TotalInputTokens() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalInputTokensLocked()
}

// TotalOutputTokens returns the total number of output tokens produced across models.
func (m *Manager) TotalOutputTokens() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalOutputTokensLocked()
}

// TotalCost calculates the total cost using accurate model-specific pricing
func (m *Manager) TotalCost() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.totalCostLocked()
}

// GetModelCost returns the cost for a specific model
func (m *Manager) GetModelCost(modelName string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	usage, exists := m.ModelUsage[modelName]
	if !exists {
		return 0.0
	}

	return m.pricing.CalculateCost(modelName, usage.InputTokens, usage.OutputTokens)
}

// GetModelUsage returns the token usage for a specific model
func (m *Manager) GetModelUsage(modelName string) TokenUsage {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.ModelUsage[modelName]
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
	if m.BudgetDollars > 0 && m.totalCostLocked() > m.BudgetDollars {
		return true
	}
	return false
}

func (m *Manager) totalCostLocked() float64 {
	totalCost := 0.0

	// Calculate cost for each model using specific pricing
	for modelName, usage := range m.ModelUsage {
		cost := m.pricing.CalculateCost(modelName, usage.InputTokens, usage.OutputTokens)
		totalCost += cost
	}
	return totalCost
}
