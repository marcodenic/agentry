package cost

import "sync"



// TokenUsage represents token usage for a model call
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

type Manager struct {
	mu            sync.Mutex
	ModelUsage    map[string]TokenUsage // Model name -> token usage
	ToolTokens    map[string]int        // Tool name -> token count (deprecated)
	BudgetTokens  int
	BudgetDollars float64
	pricing       *PricingTable
}

func New(budgetTokens int, budgetDollars float64) *Manager {
	return &Manager{
		ModelUsage:    map[string]TokenUsage{},
		ToolTokens:    map[string]int{},
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

// AddModel adds tokens for a model (deprecated, use AddModelUsage instead)
func (m *Manager) AddModel(name string, tokens int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For backward compatibility, treat all tokens as output tokens
	current := m.ModelUsage[name]
	current.OutputTokens += tokens
	m.ModelUsage[name] = current

	return m.overBudgetLocked()
}

// AddTool is deprecated - tool results are included in API response token counts
// This method is kept for backward compatibility but does nothing
func (m *Manager) AddTool(name string, tokens int) bool {
	// No-op: tool results are already included in API response token counts
	// Don't even check budget since we're not adding anything
	return false
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
	// Note: Tool tokens are no longer counted separately as they're included in API response token counts
	// The ToolTokens map is kept for backward compatibility but not used in calculations
	return total
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

	// Note: Tool costs are no longer tracked separately as they're included in API response token counts
	// The ToolTokens map is kept for backward compatibility but not used in calculations

	return totalCost
}
