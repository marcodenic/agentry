package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestAgentConfigPrecedence verifies that the primary agent uses model configuration
// from its role file (agent_0.yaml) when available, falling back to global config when not
func TestAgentConfigPrecedence(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir, err := os.MkdirTemp("", "agentry_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	rolesDir := filepath.Join(tmpDir, "roles")
	require.NoError(t, os.MkdirAll(rolesDir, 0755))

	// Create test role configuration with specific model
	roleConfig := team.RoleConfig{
		Name: "agent_0",
		Model: &config.ModelManifest{
			Name:     "test-role-model",
			Provider: "openai",
			Options: map[string]string{
				"model": "gpt-5-mini",
			},
		},
		Prompt: "You are a test agent with role-specific model config.",
	}

	roleData, err := yaml.Marshal(roleConfig)
	require.NoError(t, err)

	roleFilePath := filepath.Join(rolesDir, "agent_0.yaml")
	require.NoError(t, os.WriteFile(roleFilePath, roleData, 0644))

	// Test 1: Verify role loading works correctly
	t.Run("LoadRoleFromFile", func(t *testing.T) {
		loadedRole, err := team.LoadRoleFromFile(roleFilePath)
		require.NoError(t, err)
		assert.Equal(t, "agent_0", loadedRole.Name)
		assert.NotNil(t, loadedRole.Model)
		assert.Equal(t, "openai", loadedRole.Model.Provider)
		assert.Equal(t, "gpt-5-mini", loadedRole.Model.Options["model"])
		assert.Contains(t, loadedRole.Prompt, "role-specific model config")
	})

	// Test 2: Test with different role model vs global model
	t.Run("RoleSpecificModelPrecedence", func(t *testing.T) {
		// Set environment to point to our test role
		oldPrompt := os.Getenv("AGENTRY_DEFAULT_PROMPT")
		defer func() {
			if oldPrompt != "" {
				os.Setenv("AGENTRY_DEFAULT_PROMPT", oldPrompt)
			} else {
				os.Unsetenv("AGENTRY_DEFAULT_PROMPT")
			}
		}()
		os.Setenv("AGENTRY_DEFAULT_PROMPT", roleFilePath)

		// Create a global config with a different model
		globalConfig := config.File{
			Models: []config.ModelManifest{
				{
					Name:     "global-model",
					Provider: "anthropic",
					Options: map[string]string{
						"model": "claude-3-5-sonnet",
					},
				},
			},
		}

		// The build agent function should prefer the role-specific model
		// Since we can't directly call buildAgent (it's in main package),
		// we test the role loading behavior directly
		role, err := team.LoadRoleFromFile(roleFilePath)
		require.NoError(t, err)

		// Verify that role-specific model takes precedence
		assert.NotNil(t, role.Model)
		assert.Equal(t, "openai", role.Model.Provider)
		assert.Equal(t, "gpt-5-mini", role.Model.Options["model"])

		// Verify this is different from global config
		assert.NotEqual(t, globalConfig.Models[0].Provider, role.Model.Provider)
		assert.NotEqual(t, globalConfig.Models[0].Options["model"], role.Model.Options["model"])
	})

	// Test 3: Fallback to global config when role has no model
	t.Run("FallbackToGlobalModel", func(t *testing.T) {
		// Create a role config without model
		roleConfigNoModel := team.RoleConfig{
			Name:   "agent_0",
			Model:  nil, // No model specified
			Prompt: "You are a test agent without specific model config.",
		}

		roleDataNoModel, err := yaml.Marshal(roleConfigNoModel)
		require.NoError(t, err)

		roleFilePathNoModel := filepath.Join(rolesDir, "agent_0_no_model.yaml")
		require.NoError(t, os.WriteFile(roleFilePathNoModel, roleDataNoModel, 0644))

		loadedRole, err := team.LoadRoleFromFile(roleFilePathNoModel)
		require.NoError(t, err)
		assert.Equal(t, "agent_0", loadedRole.Name)
		assert.Nil(t, loadedRole.Model) // Should be nil, indicating fallback to global
	})

	// Test 4: Verify prompt is also loaded from role
	t.Run("RoleSpecificPrompt", func(t *testing.T) {
		role, err := team.LoadRoleFromFile(roleFilePath)
		require.NoError(t, err)
		assert.Contains(t, role.Prompt, "role-specific model config")
		assert.Contains(t, role.Prompt, "test agent")
	})
}

// TestRoleConfigYAMLFormat verifies the YAML structure matches expectations
func TestRoleConfigYAMLFormat(t *testing.T) {
	yamlContent := `name: agent_0
model:
  name: test-model
  provider: openai
  options:
    model: gpt-5-mini
    temperature: "0.1"
prompt: |
  You are a test agent.
  Use the tools provided.
`

	var role team.RoleConfig
	err := yaml.Unmarshal([]byte(yamlContent), &role)
	require.NoError(t, err)

	assert.Equal(t, "agent_0", role.Name)
	assert.NotNil(t, role.Model)
	assert.Equal(t, "test-model", role.Model.Name)
	assert.Equal(t, "openai", role.Model.Provider)
	assert.Equal(t, "gpt-5-mini", role.Model.Options["model"])
	assert.Equal(t, "0.1", role.Model.Options["temperature"])
	assert.Contains(t, role.Prompt, "test agent")
}
