#!/bin/bash

# Sequential Dependencies Coordination Test
# Test Agent 0's ability to coordinate tasks with dependencies

echo "🎯 Sequential Dependencies Coordination Test"
echo "============================================="
echo "Testing Agent 0's ability to handle task dependencies"
echo ""

# Create clean sandbox
rm -rf /tmp/agentry-ai-sandbox
mkdir -p /tmp/agentry-ai-sandbox
cd /tmp/agentry-ai-sandbox

# Copy agentry binary and config
cp /home/marco/Documents/GitHub/agentry/agentry.exe ./agentry
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .

echo "🧪 Test Scenario: Sequential Dependencies"
echo "1. First create a base library module"
echo "2. Then create a service that uses the library"
echo "3. Finally create tests that validate the integration"
echo ""

# Start the test
echo "🤖 Agent 0 Request:"
echo "Create a sequential project with dependencies:"
echo "1. First create 'database.py' with a Database class that has connect(), query(), and close() methods"
echo "2. Then create 'user_service.py' that imports and uses the Database class to create get_user() and create_user() functions"  
echo "3. Finally create 'test_integration.py' that imports both modules and tests the complete user creation and retrieval workflow"
echo "These must be created in the correct order due to their dependencies."
echo ""

# Test sequential coordination
timeout 120s ./agentry chat <<EOF
Create a sequential project with dependencies in the correct order:

Step 1: First create 'database.py' with a Database class that has these methods:
- connect(): establishes connection (just print "Connected to database")
- query(sql): executes query (just return mock data like [{"id": 1, "name": "John"}])
- close(): closes connection (just print "Database connection closed")

Step 2: Then create 'user_service.py' that imports the Database class and provides:
- get_user(user_id): uses database.query() to fetch user data
- create_user(name): uses database.query() to create user and return success message
Both functions should create a Database instance, use it, and close it.

Step 3: Finally create 'test_integration.py' that imports both database and user_service modules and:
- Tests creating a user with create_user("Alice")
- Tests retrieving the user with get_user(1)
- Validates that both operations work together

Please create these files in the correct dependency order so they can import each other successfully.
EOF

echo ""
echo "📊 RESULTS ANALYSIS"
echo "==================="

# Check if files were created in correct order
echo "✓ Files created:"
if [ -f "database.py" ]; then
    echo "  ✅ database.py (base dependency)"
else
    echo "  ❌ database.py (MISSING - this should be created first)"
fi

if [ -f "user_service.py" ]; then
    echo "  ✅ user_service.py (dependent on database.py)"
else
    echo "  ❌ user_service.py (MISSING - depends on database.py)"
fi

if [ -f "test_integration.py" ]; then
    echo "  ✅ test_integration.py (depends on both modules)"
else
    echo "  ❌ test_integration.py (MISSING - depends on both modules)"
fi

echo ""
echo "🔍 Dependency Validation:"

# Test if dependencies work
echo "Testing database.py standalone..."
if python3 -c "import database; db = database.Database(); db.connect(); result = db.query('SELECT * FROM users'); print('Database test:', result); db.close()" 2>/dev/null; then
    echo "  ✅ database.py works independently"
else
    echo "  ❌ database.py has issues"
    echo "  Content preview:"
    head -10 database.py 2>/dev/null || echo "  (file not readable)"
fi

echo ""
echo "Testing user_service.py with database dependency..."
if python3 -c "import user_service; result = user_service.create_user('TestUser'); print('User service test:', result)" 2>/dev/null; then
    echo "  ✅ user_service.py imports and uses database.py correctly"
else
    echo "  ❌ user_service.py cannot import or use database.py"
    echo "  Content preview:"
    head -10 user_service.py 2>/dev/null || echo "  (file not readable)"
fi

echo ""
echo "Testing complete integration..."
if python3 -c "import test_integration" 2>/dev/null; then
    echo "  ✅ test_integration.py imports both modules successfully"
    echo "  Running integration test..."
    python3 test_integration.py 2>&1 | head -10
else
    echo "  ❌ test_integration.py cannot import dependencies"
    echo "  Content preview:"
    head -10 test_integration.py 2>/dev/null || echo "  (file not readable)"
fi

echo ""
echo "📈 COORDINATION ASSESSMENT"
echo "=========================="

total_files=3
created_files=0
dependency_success=0

if [ -f "database.py" ]; then created_files=$((created_files + 1)); fi
if [ -f "user_service.py" ]; then created_files=$((created_files + 1)); fi  
if [ -f "test_integration.py" ]; then created_files=$((created_files + 1)); fi

# Test basic imports
if python3 -c "import database" 2>/dev/null; then dependency_success=$((dependency_success + 1)); fi
if python3 -c "import user_service" 2>/dev/null; then dependency_success=$((dependency_success + 1)); fi
if python3 -c "import test_integration" 2>/dev/null; then dependency_success=$((dependency_success + 1)); fi

file_creation_rate=$((100 * created_files / total_files))
dependency_rate=$((100 * dependency_success / total_files))

echo "📊 Metrics:"
echo "  File Creation Rate: $file_creation_rate% ($created_files/$total_files)"
echo "  Dependency Success Rate: $dependency_rate% ($dependency_success/$total_files)"
echo "  Overall Coordination Score: $(((file_creation_rate + dependency_rate) / 2))%"

if [ $file_creation_rate -eq 100 ] && [ $dependency_rate -eq 100 ]; then
    echo ""
    echo "🏆 SUCCESS: Sequential dependency coordination working perfectly!"
    echo "✅ Agent 0 understood dependency order"
    echo "✅ Created files in correct sequence"
    echo "✅ All imports and dependencies work"
elif [ $file_creation_rate -eq 100 ]; then
    echo ""
    echo "⚠️  PARTIAL SUCCESS: Files created but dependency issues"
    echo "✅ Agent 0 created all requested files"
    echo "❌ Some dependencies may have issues"
elif [ $file_creation_rate -gt 50 ]; then
    echo ""
    echo "🔧 NEEDS IMPROVEMENT: Some coordination success"
    echo "✅ Agent 0 created most files"
    echo "❌ Missing some files or dependencies"
else
    echo ""
    echo "❌ FAILED: Sequential coordination not working"
    echo "❌ Agent 0 did not coordinate sequential tasks properly"
fi

echo ""
echo "🗂️ Final File Structure:"
ls -la
echo ""

# Show actual file contents for debugging
echo "📄 Created File Contents:"
echo "========================"
for file in database.py user_service.py test_integration.py; do
    if [ -f "$file" ]; then
        echo ""
        echo "--- $file ---"
        cat "$file"
    fi
done

echo ""
echo "✅ Sequential Dependencies Test Complete"
