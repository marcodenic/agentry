#!/bin/bash

# Complex Multi-Agent Orchestration Test
# Test Agent 0's ability to manage a team through complex, multi-step coordination

# Source the test helpers script
# shellcheck source=/dev/null
source "$(dirname "$0")/../scripts/test-helpers.sh"

echo "🎭 Complex Multi-Agent Orchestration Test"
echo "========================================="
echo "Testing Agent 0's full orchestration capabilities:"
echo "- Spawn multiple agents"
echo "- Delegate complex, interdependent tasks" 
echo "- Monitor progress and coordinate handoffs"
echo "- Make decisions based on agent updates"
echo "- Adapt coordination strategy as needed"
echo ""

# Create clean sandbox
setup_test_environment

echo "🎯 COMPLEX ORCHESTRATION SCENARIO"
echo "================================="
echo "Build a complete e-commerce platform with multiple components:"
echo "1. Database design and schema (requires database specialist)"
echo "2. Backend API with authentication and product management (requires backend developer)"
echo "3. Frontend user interface with shopping cart (requires frontend developer)"  
echo "4. Payment integration service (requires specialized coder)"
echo "5. Deployment configuration and monitoring (requires devops)"
echo "6. Testing and quality assurance (requires tester)"
echo "7. Documentation and user guides (requires technical writer)"
echo ""
echo "This requires Agent 0 to:"
echo "- Check what agents are available"
echo "- Plan the work order and dependencies"
echo "- Assign tasks to appropriate specialists"
echo "- Monitor progress and coordinate handoffs"
echo "- Handle any issues or blockers"
echo "- Ensure integration between components"
echo ""

# Start the complex orchestration test
echo "🤖 Starting Complex Multi-Agent Orchestration:"
echo "=============================================="

$AGENT_CMD chat <<EOF
I need you to orchestrate the development of a complete e-commerce platform. This is a complex project that will require multiple specialized agents working together in a coordinated manner.

PROJECT GOAL: Build "ShopFlow" - a complete e-commerce platform

REQUIREMENTS:
1. Database: User accounts, product catalog, orders, inventory tracking
2. Backend API: Authentication, product management, order processing, inventory updates
3. Frontend: Product browsing, user registration/login, shopping cart, checkout process
4. Payment: Integration with payment processing (simulate with mock service)
5. Deployment: Docker configuration, basic monitoring setup
6. Testing: Unit tests for critical components, integration tests
7. Documentation: API docs, user guide, deployment instructions

ORCHESTRATION REQUIREMENTS:
- Use team_status to check available agents
- Use check_agent to verify specific agents are ready
- Plan the work order considering dependencies (database first, then API, then frontend, etc.)
- Use assign_task or agent delegation to give specific tasks to appropriate specialists
- Use send_message to coordinate between agents when needed
- Monitor progress and make decisions about next steps
- Handle any issues or conflicts that arise
- Ensure all components integrate properly

COORDINATION STRATEGY:
1. First assess your team and plan the approach
2. Start with foundational components (database schema)
3. Build core services (backend API)
4. Develop user interface (frontend)
5. Add specialized features (payment, deployment, testing)
6. Integrate and validate everything works together
7. Complete with documentation

Please show me your full orchestration process - I want to see how you manage the team, delegate tasks, coordinate between agents, and ensure the project comes together successfully.

Start by checking your team status and planning your approach.
EOF

echo ""
echo "📊 ORCHESTRATION ANALYSIS"
echo "========================="

# Wait a moment for any background processes
sleep 2

echo "🔍 Checking what was created..."
echo "Files created:"
find . -name "*.py" -o -name "*.html" -o -name "*.js" -o -name "*.sql" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" -o -name "*.json" | sort

echo ""
echo "📁 Project structure:"
ls -la

echo ""
echo "🔍 Looking for orchestration evidence..."

# Check for database files
db_files=$(find . -name "*database*" -o -name "*schema*" -o -name "*.sql" | wc -l)
echo "Database components: $db_files files"

# Check for backend files  
backend_files=$(find . -name "*api*" -o -name "*backend*" -o -name "*server*" | grep -v __pycache__ | wc -l)
echo "Backend components: $backend_files files"

# Check for frontend files
frontend_files=$(find . -name "*.html" -o -name "*.js" -o -name "*.css" | wc -l)
echo "Frontend components: $frontend_files files"

# Check for deployment files
deploy_files=$(find . -name "*docker*" -o -name "*.yaml" -o -name "*.yml" | wc -l)
echo "Deployment components: $deploy_files files"

# Check for testing files
test_files=$(find . -name "*test*" | wc -l)
echo "Testing components: $test_files files"

# Check for documentation
doc_files=$(find . -name "*.md" -o -name "*README*" -o -name "*doc*" | wc -l)
echo "Documentation components: $doc_files files"

total_files=$(find . -type f -name "*.py" -o -name "*.html" -o -name "*.js" -o -name "*.sql" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" -o -name "*.json" | wc -l)
echo "Total project files: $total_files"

echo ""
echo "🧠 ORCHESTRATION INTELLIGENCE ASSESSMENT"
echo "========================================"

# Assess coordination complexity
coordination_score=0
max_coordination_score=10

# Check if multiple component types were created (shows understanding of complex project)
if [ $db_files -gt 0 ]; then coordination_score=$((coordination_score + 1)); echo "✅ Database components created"; fi
if [ $backend_files -gt 0 ]; then coordination_score=$((coordination_score + 2)); echo "✅ Backend components created"; fi
if [ $frontend_files -gt 0 ]; then coordination_score=$((coordination_score + 2)); echo "✅ Frontend components created"; fi
if [ $deploy_files -gt 0 ]; then coordination_score=$((coordination_score + 1)); echo "✅ Deployment components created"; fi
if [ $test_files -gt 0 ]; then coordination_score=$((coordination_score + 2)); echo "✅ Testing components created"; fi
if [ $doc_files -gt 0 ]; then coordination_score=$((coordination_score + 1)); echo "✅ Documentation created"; fi
if [ $total_files -gt 10 ]; then coordination_score=$((coordination_score + 1)); echo "✅ Complex project scale (10+ files)"; fi

# Check for integration evidence
integration_score=0
max_integration_score=5

# Look for evidence of integration planning
if find . -name "*.py" -exec grep -l "database\|db\|sql" {} \; | head -1 >/dev/null 2>&1; then
    integration_score=$((integration_score + 1))
    echo "✅ Database integration evidence found"
fi

if find . -name "*.py" -exec grep -l "api\|endpoint\|route" {} \; | head -1 >/dev/null 2>&1; then
    integration_score=$((integration_score + 1)) 
    echo "✅ API integration evidence found"
fi

if find . -name "*.html" -o -name "*.js" -exec grep -l "api\|fetch\|ajax" {} \; | head -1 >/dev/null 2>&1; then
    integration_score=$((integration_score + 1))
    echo "✅ Frontend-backend integration evidence found"
fi

if find . -name "*docker*" -o -name "*.yaml" | head -1 >/dev/null 2>&1; then
    integration_score=$((integration_score + 1))
    echo "✅ Deployment integration evidence found"  
fi

if find . -name "*test*" -exec grep -l "import\|require" {} \; | head -1 >/dev/null 2>&1; then
    integration_score=$((integration_score + 1))
    echo "✅ Testing integration evidence found"
fi

# Calculate final scores
coordination_percentage=$((100 * coordination_score / max_coordination_score))
integration_percentage=$((100 * integration_score / max_integration_score))
overall_orchestration=$((60 * coordination_percentage / 100 + 40 * integration_percentage / 100))

echo ""
echo "📈 ORCHESTRATION METRICS:"
echo "========================"
echo "Component Coordination: $coordination_percentage% ($coordination_score/$max_coordination_score)"
echo "Integration Planning: $integration_percentage% ($integration_score/$max_integration_score)" 
echo "Overall Orchestration Score: $overall_orchestration%"

echo ""
echo "🎯 ORCHESTRATION ASSESSMENT:"
echo "============================"

if [ $overall_orchestration -ge 90 ]; then
    echo "🏆 OUTSTANDING: Agent 0 demonstrates MASTER-LEVEL orchestration!"
    echo "✅ Complex multi-component project coordination"
    echo "✅ Proper integration planning and execution"
    echo "✅ Full-stack development orchestration"
elif [ $overall_orchestration -ge 75 ]; then
    echo "🎉 EXCELLENT: Agent 0 shows strong orchestration capabilities!"
    echo "✅ Good multi-component coordination"
    echo "✅ Most integration aspects handled well"
    echo "⚠️  Minor areas for orchestration improvement"
elif [ $overall_orchestration -ge 60 ]; then
    echo "✅ GOOD: Agent 0 handles complex coordination reasonably well"
    echo "✅ Basic multi-component understanding"
    echo "⚠️  Integration coordination needs improvement"
else
    echo "⚠️  NEEDS IMPROVEMENT: Complex orchestration capabilities need development"
    echo "❌ Limited multi-component coordination"
    echo "❌ Integration planning needs significant work"
fi

echo ""
echo "📄 SAMPLE ORCHESTRATION EVIDENCE:"
echo "================================="

# Show evidence of complex coordination
echo "🗂️ Project Structure Created:"
find . -type f \( -name "*.py" -o -name "*.html" -o -name "*.js" -o -name "*.sql" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" \) -exec echo "  {}" \;

echo ""
echo "📋 Key Integration Files:"
for file in $(find . -name "*api*" -o -name "*main*" -o -name "*app*" | head -3); do
    if [ -f "$file" ]; then
        echo ""
        echo "--- $file ---"
        head -10 "$file"
    fi
done

echo ""
echo "✅ Complex Multi-Agent Orchestration Test Complete"
echo ""
echo "🔄 NEXT STEPS FOR VALIDATION:"
echo "- Review trace output for evidence of agent delegation"
echo "- Check for team_status, check_agent, assign_task tool usage" 
echo "- Verify multi-agent communication patterns"
echo "- Validate complex project coordination vs single-agent execution"
