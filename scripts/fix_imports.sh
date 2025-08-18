#!/bin/bash

# Script to fix converse imports across the project

files_with_converse=(
    "./examples/test-programs/test-programs/test_delegation_debug.go"
    "./examples/test-programs/test-programs/test_final_verification.go"
    "./examples/test-programs/test_final_verification.go"
    "./examples/test-programs/test_full_workflow.go"
    "./examples/test_delegation_scenario.go"
    "./tests/builtin_cross_test.go"
    # cleaned: removed obsolete user_prompt_debug_test, debug helpers
    "./tests/agent_tool_context_test.go"
    "./tests/cross_platform_agent_workflow_test.go"
    "./tests/converse_team_integration_test.go"
    "./tests/new_team_context_test.go"
    "./tests/tui_agent_tool_test.go"
    "./tests/user_prompt_realistic_test.go"
    "./tests/semantic_commands_test.go"
    "./tests/converse_test.go"
    "./tests/multi_agent_role_test.go"
    "./tests/complete_workflow_test.go"
)

for file in "${files_with_converse[@]}"; do
    if [ -f "$file" ]; then
        echo "Fixing $file"
        # Replace the import
        sed -i 's|"github.com/marcodenic/agentry/internal/converse"|"github.com/marcodenic/agentry/internal/team"|g' "$file"
        # Replace function calls
        sed -i 's|converse\.NewTeam|team.NewTeam|g' "$file"
        sed -i 's|converse\.|team.|g' "$file"
    fi
done

echo "Fixed converse imports in all files"
