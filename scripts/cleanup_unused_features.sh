#!/bin/bash

# Cleanup script to remove unused/unfinished features from Agentry
# Focus on the working core: Agent 0 delegation, TUI, and basic tools

set -e

echo "🧹 Cleaning up unused features from Agentry..."

# Remove persistent mode infrastructure
echo "📁 Removing persistent mode infrastructure..."
rm -rf internal/sessions/
rm -f config/persistent-config.yaml
rm -f config/simple-session-config.yaml
rm -f config/session-test-config.yaml

# Remove NATS/message queue infrastructure  
echo "📡 Removing NATS/message queue infrastructure..."
rm -rf internal/taskqueue/
rm -rf internal/mocknats/
rm -f cmd/autoscaler/main.go
rm -f cmd/worker/main.go
rmdir cmd/autoscaler/ 2>/dev/null || true
rmdir cmd/worker/ 2>/dev/null || true

# Remove Kubernetes deployment
echo "☸️  Removing Kubernetes deployment..."
rm -rf deploy/k8s/
rm -rf deploy/helm/

# Remove autoscaling
echo "📈 Removing autoscaling infrastructure..."
rm -rf pkg/autoscale/

# Remove VS Code extension
echo "🔧 Removing VS Code extension..."
rm -rf extensions/vscode-agentry/

# Remove unused documentation about unfinished features
echo "📚 Cleaning up documentation..."
rm -f PLAN_DETAILED.md
rm -f PLAN_CONDENSED.md
rm -f COLLABORATION_SUCCESS.md
rm -f CONFIG_CLEANUP_SUMMARY.md
rm -f REGISTRY_ARCHITECTURE_DECISION.md
rm -f GPT41_NANO_UPDATE.md

# Remove test files for removed features
echo "🧪 Removing tests for removed features..."
find tests/ -name "*persistent*" -delete 2>/dev/null || true
find tests/ -name "*session*" -delete 2>/dev/null || true
find tests/ -name "*coordination*" -delete 2>/dev/null || true
rm -rf tests/coordination/ 2>/dev/null || true

# Clean up go.mod dependencies (will need manual review)
echo "📦 Note: You'll need to manually clean up go.mod dependencies"

echo "✅ Cleanup complete!"
echo ""
echo "🎯 Agentry is now focused on:"
echo "   - Agent 0 delegation system (working)"
echo "   - TUI interface (working)" 
echo "   - Advanced file operations (working)"
echo "   - Basic multi-agent coordination (working)"
echo "   - Cross-platform tools (working)"
echo ""
echo "🚮 Removed:"
echo "   - Persistent mode / sessions"
echo "   - NATS message queues"
echo "   - Kubernetes deployment"
echo "   - Autoscaling"
echo "   - VS Code extension"
echo "   - Complex coordination docs"
echo ""
echo "📝 Next steps:"
echo "   1. Update go.mod to remove unused dependencies"
echo "   2. Update README.md to reflect simplified focus"
echo "   3. Update .agentry.yaml configs"
echo "   4. Test the core functionality"
