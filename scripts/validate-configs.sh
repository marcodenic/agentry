#!/bin/bash

# Config validation script for Agentry
# This script checks all YAML config files for common issues

echo "🔍 Validating Agentry configuration files..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to check a config file
check_config() {
    local file="$1"
    echo -e "\n📄 Checking: $file"
    
    # Check if file exists
    if [[ ! -f "$file" ]]; then
        echo -e "${RED}❌ File not found${NC}"
        return 1
    fi
    
    # Check YAML syntax
    if ! python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>/dev/null; then
        echo -e "${RED}❌ Invalid YAML syntax${NC}"
        python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>&1 | head -3
        return 1
    fi
    
    # Check for old routing configuration
    if grep -q "routes:" "$file" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  Contains legacy 'routes' configuration${NC}"
    fi
    
    if grep -q "if_contains:" "$file" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  Contains legacy 'if_contains' configuration${NC}"
    fi
    
    # Check for old model format
    if grep -q "^model:" "$file" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  Uses old 'model:' format instead of 'models:' array${NC}"
    fi
    
    # Check for agent tool (required for delegation)
    if grep -q "tools:" "$file" 2>/dev/null; then
        if ! grep -q "name: agent" "$file" 2>/dev/null; then
            echo -e "${YELLOW}⚠️  Missing 'agent' tool (required for delegation)${NC}"
        fi
    fi
    
    # Check for hardcoded API keys
    if grep -q "apiKey:" "$file" 2>/dev/null; then
        if ! grep -q "\${" "$file" 2>/dev/null; then
            echo -e "${YELLOW}⚠️  May contain hardcoded API key${NC}"
        fi
    fi
    
    echo -e "${GREEN}✅ Basic validation passed${NC}"
}

# Main config files
echo -e "\n📂 Main Configuration Files:"
check_config ".agentry.yaml"
check_config "examples/.agentry.yaml"

# Test config files
echo -e "\n📂 Test Configuration Files:"
for config in config/*.yaml; do
    if [[ -f "$config" ]]; then
        check_config "$config"
    fi
done

# Specialized test configs
echo -e "\n📂 Specialized Test Configurations:"
for config in tests/**/*.yaml; do
    if [[ -f "$config" ]]; then
        check_config "$config"
    fi
done

echo -e "\n✨ Configuration validation complete!"
echo -e "\nFor detailed configuration information, see: CONFIG_GUIDE.md"
