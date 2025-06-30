package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileRegistry implements AgentRegistry using a JSON file for cross-platform persistence
type FileRegistry struct {
	agents     map[string]*AgentInfo
	health     map[string]*HealthMetrics
	mutex      sync.RWMutex
	configFile string
	events     []EventSubscriber
}

// RegistryConfig holds the configuration for the file registry
type RegistryConfig struct {
	ConfigDir string `json:"config_dir"`
	FileName  string `json:"file_name"`
}

// NewFileRegistry creates a new file-based agent registry
func NewFileRegistry(config *RegistryConfig) (*FileRegistry, error) {
	if config == nil {
		config = DefaultRegistryConfig()
	}

	// Ensure config directory exists
	if err := os.MkdirAll(config.ConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(config.ConfigDir, config.FileName)
	
	registry := &FileRegistry{
		agents:     make(map[string]*AgentInfo),
		health:     make(map[string]*HealthMetrics),
		configFile: configFile,
		events:     make([]EventSubscriber, 0),
	}

	// Load existing agents from file
	if err := registry.loadFromFile(); err != nil {
		// If file doesn't exist, that's ok - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load existing registry: %w", err)
		}
	}

	return registry, nil
}

// DefaultRegistryConfig returns the default configuration for the file registry
func DefaultRegistryConfig() *RegistryConfig {
	// Use cross-platform temporary directory
	tempDir := os.TempDir()
	agentryDir := filepath.Join(tempDir, "agentry")
	
	return &RegistryConfig{
		ConfigDir: agentryDir,
		FileName:  "agents.json",
	}
}

// Close closes the registry and any underlying resources
func (r *FileRegistry) Close() error {
	// Save final state
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.saveToFile()
}

// hasCapabilities checks if an agent has all required capabilities
func (r *FileRegistry) hasCapabilities(agent *AgentInfo, required []string) bool {
	agentCaps := make(map[string]bool)
	for _, cap := range agent.Capabilities {
		agentCaps[cap] = true
	}

	for _, reqCap := range required {
		if !agentCaps[reqCap] {
			return false
		}
	}

	return true
}
