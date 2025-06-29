#!/bin/bash

# Mixed Coordination Test: Parallel + Sequential Dependencies
# Test Agent 0's ability to handle complex coordination with both parallel and sequential tasks

echo "🎭 Mixed Coordination Test: Parallel + Sequential"  
echo "================================================="
echo "Testing Agent 0's advanced coordination intelligence"
echo ""

# Create clean sandbox
rm -rf /tmp/agentry-ai-sandbox
mkdir -p /tmp/agentry-ai-sandbox
cd /tmp/agentry-ai-sandbox

# Copy agentry binary and config
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "🧪 Test Scenario: Mixed Parallel + Sequential Coordination"
echo "Complex project with both parallel and sequential dependencies:"
echo "1. PARALLEL: Create 3 independent utility modules (math, string, date utils)"
echo "2. SEQUENTIAL: Create main service that depends on ALL 3 utils"
echo "3. SEQUENTIAL: Create test suite that depends on main service"
echo "4. PARALLEL: Create documentation for each component"
echo ""

# Start the complex coordination test
echo "🤖 Agent 0 Request:"
echo "Create a complex project with mixed parallel and sequential coordination:"
echo ""

timeout 180s ./agentry chat <<EOF
Create a complex project structure with mixed coordination patterns:

PHASE 1 - PARALLEL TASKS (these can be done simultaneously):
- Create 'math_utils.py' with functions: add(a,b), multiply(a,b), factorial(n)
- Create 'string_utils.py' with functions: reverse_string(s), capitalize_words(s), count_vowels(s)  
- Create 'date_utils.py' with functions: current_timestamp(), days_between(date1, date2), format_date(date)

PHASE 2 - SEQUENTIAL DEPENDENCY (requires Phase 1 complete):
- Create 'main_service.py' that imports and uses ALL three utility modules to provide:
  * process_data(text, numbers, date) - uses all three utils
  * generate_report() - creates a report using all utilities

PHASE 3 - SEQUENTIAL DEPENDENCY (requires Phase 2 complete):
- Create 'test_main_service.py' that imports main_service and tests:
  * Test process_data() with sample inputs
  * Test generate_report()
  * Verify integration of all utility modules

PHASE 4 - PARALLEL TASKS (these can be done simultaneously after Phase 3):
- Create 'README.md' with project overview and usage examples
- Create 'ARCHITECTURE.md' with component descriptions and dependencies
- Create 'TESTING.md' with testing instructions and examples

This project tests your ability to understand and coordinate both parallel and sequential dependencies correctly.
EOF

echo ""
echo "📊 MIXED COORDINATION ANALYSIS"
echo "==============================="

# Count files in each phase
echo "✅ PHASE 1 - Parallel Utilities:"
phase1_count=0
if [ -f "math_utils.py" ]; then echo "  ✅ math_utils.py"; phase1_count=$((phase1_count + 1)); else echo "  ❌ math_utils.py (missing)"; fi
if [ -f "string_utils.py" ]; then echo "  ✅ string_utils.py"; phase1_count=$((phase1_count + 1)); else echo "  ❌ string_utils.py (missing)"; fi
if [ -f "date_utils.py" ]; then echo "  ✅ date_utils.py"; phase1_count=$((phase1_count + 1)); else echo "  ❌ date_utils.py (missing)"; fi

echo ""
echo "🔗 PHASE 2 - Sequential Main Service:"
phase2_count=0
if [ -f "main_service.py" ]; then echo "  ✅ main_service.py"; phase2_count=$((phase2_count + 1)); else echo "  ❌ main_service.py (missing)"; fi

echo ""
echo "🧪 PHASE 3 - Sequential Test Suite:"
phase3_count=0
if [ -f "test_main_service.py" ]; then echo "  ✅ test_main_service.py"; phase3_count=$((phase3_count + 1)); else echo "  ❌ test_main_service.py (missing)"; fi

echo ""
echo "📚 PHASE 4 - Parallel Documentation:"
phase4_count=0
if [ -f "README.md" ]; then echo "  ✅ README.md"; phase4_count=$((phase4_count + 1)); else echo "  ❌ README.md (missing)"; fi
if [ -f "ARCHITECTURE.md" ]; then echo "  ✅ ARCHITECTURE.md"; phase4_count=$((phase4_count + 1)); else echo "  ❌ ARCHITECTURE.md (missing)"; fi
if [ -f "TESTING.md" ]; then echo "  ✅ TESTING.md"; phase4_count=$((phase4_count + 1)); else echo "  ❌ TESTING.md (missing)"; fi

echo ""
echo "🔍 DEPENDENCY VALIDATION"
echo "========================"

# Test Phase 1 utilities work independently (parallel requirement)
echo "Testing Phase 1 utilities (should work independently):"
for util in math_utils string_utils date_utils; do
    if python3 -c "import $util; print('✅ $util imports successfully')" 2>/dev/null; then
        echo "  ✅ $util.py works independently"
    else
        echo "  ❌ $util.py has issues"
    fi
done

echo ""
echo "Testing Phase 2 integration (sequential dependency on Phase 1):"
if [ -f "main_service.py" ]; then
    if python3 -c "import main_service; print('✅ main_service imports successfully')" 2>/dev/null; then
        echo "  ✅ main_service.py imports all Phase 1 dependencies correctly"
        
        # Check if it actually uses the utilities
        echo "  🔍 Checking Phase 1 dependency usage in main_service.py:"
        if grep -q "import.*math_utils\|from.*math_utils" main_service.py; then echo "    ✅ Uses math_utils"; else echo "    ❌ Missing math_utils usage"; fi
        if grep -q "import.*string_utils\|from.*string_utils" main_service.py; then echo "    ✅ Uses string_utils"; else echo "    ❌ Missing string_utils usage"; fi
        if grep -q "import.*date_utils\|from.*date_utils" main_service.py; then echo "    ✅ Uses date_utils"; else echo "    ❌ Missing date_utils usage"; fi
    else
        echo "  ❌ main_service.py cannot import Phase 1 dependencies"
    fi
else
    echo "  ❌ main_service.py not created"
fi

echo ""
echo "Testing Phase 3 integration (sequential dependency on Phase 2):"
if [ -f "test_main_service.py" ]; then
    if python3 -c "import test_main_service" 2>/dev/null; then
        echo "  ✅ test_main_service.py imports main_service correctly"
        
        # Try to run the test
        echo "  🧪 Running integration test:"
        python3 test_main_service.py 2>&1 | head -10
    else
        echo "  ❌ test_main_service.py cannot import dependencies"
    fi
else
    echo "  ❌ test_main_service.py not created"
fi

echo ""
echo "📈 COORDINATION INTELLIGENCE ASSESSMENT"
echo "========================================"

# Calculate coordination scores
total_phase1=3
total_phase2=1  
total_phase3=1
total_phase4=3

phase1_success_rate=$((100 * phase1_count / total_phase1))
phase2_success_rate=$((100 * phase2_count / total_phase2))
phase3_success_rate=$((100 * phase3_count / total_phase3))
phase4_success_rate=$((100 * phase4_count / total_phase4))

overall_file_rate=$(((100 * (phase1_count + phase2_count + phase3_count + phase4_count)) / (total_phase1 + total_phase2 + total_phase3 + total_phase4)))

echo "📊 Phase Success Rates:"
echo "  Phase 1 (Parallel Utils): $phase1_success_rate% ($phase1_count/$total_phase1)"
echo "  Phase 2 (Sequential Service): $phase2_success_rate% ($phase2_count/$total_phase2)"  
echo "  Phase 3 (Sequential Tests): $phase3_success_rate% ($phase3_count/$total_phase3)"
echo "  Phase 4 (Parallel Docs): $phase4_success_rate% ($phase4_count/$total_phase4)"
echo "  Overall File Creation: $overall_file_rate%"

# Test dependency understanding
dependency_score=0
max_dependency_score=6

# Phase 1 utilities should work independently
for util in math_utils string_utils date_utils; do
    if python3 -c "import $util" 2>/dev/null; then dependency_score=$((dependency_score + 1)); fi
done

# Phase 2 should depend on Phase 1
if [ -f "main_service.py" ] && python3 -c "import main_service" 2>/dev/null; then dependency_score=$((dependency_score + 1)); fi

# Phase 3 should depend on Phase 2  
if [ -f "test_main_service.py" ] && python3 -c "import test_main_service" 2>/dev/null; then dependency_score=$((dependency_score + 1)); fi

# Phase 4 can be independent
if [ -f "README.md" ]; then dependency_score=$((dependency_score + 1)); fi

dependency_understanding=$((100 * dependency_score / max_dependency_score))

echo "  Dependency Understanding: $dependency_understanding% ($dependency_score/$max_dependency_score)"

# Overall coordination intelligence score
coordination_intelligence=$(((overall_file_rate + dependency_understanding) / 2))

echo ""
echo "🎯 FINAL COORDINATION ASSESSMENT:"
echo "================================="
echo "🧠 Coordination Intelligence Score: $coordination_intelligence%"

if [ $coordination_intelligence -ge 90 ]; then
    echo ""
    echo "🏆 OUTSTANDING: Agent 0 demonstrates ADVANCED coordination intelligence!"
    echo "✅ Perfect understanding of parallel vs sequential dependencies"
    echo "✅ Proper task ordering and dependency management"
    echo "✅ High-quality implementation across all phases"
elif [ $coordination_intelligence -ge 75 ]; then
    echo ""
    echo "🎉 EXCELLENT: Agent 0 shows strong coordination capabilities!"
    echo "✅ Good understanding of mixed coordination patterns"
    echo "✅ Most dependencies handled correctly"
    echo "⚠️  Minor areas for improvement"
elif [ $coordination_intelligence -ge 60 ]; then
    echo ""
    echo "✅ GOOD: Agent 0 handles basic mixed coordination"
    echo "✅ Some parallel/sequential understanding"
    echo "❌ Room for improvement in complex dependencies"
else
    echo ""
    echo "⚠️  NEEDS IMPROVEMENT: Mixed coordination capabilities need work"
    echo "❌ Limited understanding of parallel vs sequential patterns"
fi

echo ""
echo "🗂️ Final Project Structure:"
ls -la

echo ""
echo "📄 Sample File Contents (showing coordination quality):"
echo "======================================================="

# Show key integration files
if [ -f "main_service.py" ]; then
    echo ""
    echo "--- main_service.py (shows Phase 1→2 dependency) ---"
    cat main_service.py
fi

if [ -f "test_main_service.py" ]; then
    echo ""
    echo "--- test_main_service.py (shows Phase 2→3 dependency) ---"
    head -15 test_main_service.py
fi

echo ""
echo "✅ Mixed Coordination Test Complete"
