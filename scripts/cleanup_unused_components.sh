#!/bin/bash

echo "🧹 Cleaning up unused Agentry components..."

# Remove unused internal modules
echo "Removing unused internal modules..."
rm -rf internal/collaboration/
rm -rf internal/plugin/
rm -rf internal/policy/

# Remove agentry-demos (these are just examples)
echo "Removing demo workflows..."
rm -rf agentry-demos/

# Remove plugin/policy related test files
echo "Removing related test files..."
rm -f tests/plugin_*_test.go
rm -f tests/policy_test.go

echo "✅ Cleanup complete!"
echo ""
echo "The following components were removed:"
echo "- internal/collaboration/ - Multi-agent collaboration engine (unused)"
echo "- internal/plugin/ - Plugin management system (stub functions only)"
echo "- internal/policy/ - Tool approval/security policies (server-only feature)"
echo "- agentry-demos/ - Demo workflows (not core functionality)"
echo ""
echo "Router was kept because it's used by the core agent for model selection."
