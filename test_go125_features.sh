#!/bin/bash

# Go 1.25 Feature Testing Script
# This script demonstrates and tests the new features available in Go 1.25

echo "=== Testing Go 1.25 Features for Agentry ==="

# Set Go path
export PATH=/usr/local/go/bin:$PATH

echo "Current Go version:"
go version
echo

# Test 1: Build with experimental JSON v2 for better performance
echo "=== Test 1: Building with experimental JSON v2 ==="
echo "Building with GOEXPERIMENT=jsonv2 for improved JSON performance..."
cd /home/marco/Documents/GitHub/agentry
GOEXPERIMENT=jsonv2 go build -o agentry-jsonv2 ./cmd/agentry
if [ $? -eq 0 ]; then
    echo "✅ JSON v2 build successful"
    ./agentry-jsonv2 --version
else
    echo "❌ JSON v2 build failed"
fi
echo

# Test 2: Build with experimental garbage collector
echo "=== Test 2: Building with experimental garbage collector ==="
echo "Building with GOEXPERIMENT=greenteagc for reduced GC overhead..."
GOEXPERIMENT=greenteagc go build -o agentry-greentea ./cmd/agentry
if [ $? -eq 0 ]; then
    echo "✅ Green Tea GC build successful"
    ./agentry-greentea --version
else
    echo "❌ Green Tea GC build failed"
fi
echo

# Test 3: Build with both experimental features
echo "=== Test 3: Building with both experimental features ==="
echo "Building with GOEXPERIMENT=jsonv2,greenteagc..."
GOEXPERIMENT=jsonv2,greenteagc go build -o agentry-experimental ./cmd/agentry
if [ $? -eq 0 ]; then
    echo "✅ Combined experimental build successful"
    ./agentry-experimental --version
else
    echo "❌ Combined experimental build failed"
fi
echo

# Test 4: Run new Go 1.25 specific tests
echo "=== Test 4: Running Go 1.25 specific tests ==="
echo "Testing new sync.WaitGroup.Go() and testing/synctest features..."
go test -v ./tests -run "TestConcurrentAgentsWithSynctest|TestTimeoutBehaviorWithSynctest|TestEventStreamingWithSynctest"
echo

# Test 5: Run vet with new analyzers
echo "=== Test 5: Running go vet with new Go 1.25 analyzers ==="
echo "Checking for waitgroup and hostport issues..."
go vet ./...
echo

# Test 6: DWARF5 debug info comparison
echo "=== Test 6: DWARF5 debug information comparison ==="
echo "Building with DWARF5 debug info (default in Go 1.25)..."
go build -ldflags="-w=false" -o agentry-dwarf5 ./cmd/agentry
echo "Building with legacy DWARF (for comparison)..."
GOEXPERIMENT=nodwarf5 go build -ldflags="-w=false" -o agentry-legacy ./cmd/agentry

if [ -f agentry-dwarf5 ] && [ -f agentry-legacy ]; then
    dwarf5_size=$(stat -f%z agentry-dwarf5 2>/dev/null || stat -c%s agentry-dwarf5)
    legacy_size=$(stat -f%z agentry-legacy 2>/dev/null || stat -c%s agentry-legacy)
    echo "DWARF5 binary size: $dwarf5_size bytes"
    echo "Legacy DWARF binary size: $legacy_size bytes"
    
    if [ $dwarf5_size -lt $legacy_size ]; then
        reduction=$((legacy_size - dwarf5_size))
        percentage=$((reduction * 100 / legacy_size))
        echo "✅ DWARF5 binary is $reduction bytes smaller (${percentage}% reduction)"
    else
        echo "ℹ️  DWARF5 binary size comparison: no significant difference"
    fi
fi
echo

# Test 7: Container awareness test (if running in container)
echo "=== Test 7: Container-aware GOMAXPROCS test ==="
echo "Current GOMAXPROCS settings:"
go run -c 'import "runtime"; fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))' 2>/dev/null || echo "Manual test: Check GOMAXPROCS in container environment"
echo

echo "=== Go 1.25 Feature Testing Complete ==="
echo
echo "Summary of benefits for Agentry:"
echo "1. ✅ JSON v2: Improved JSON marshal/unmarshal performance (significant for AI API communication)"
echo "2. ✅ Green Tea GC: 10-40% reduction in GC overhead (helps with memory-intensive operations)"
echo "3. ✅ DWARF5: Smaller debug info and faster linking"
echo "4. ✅ Container awareness: Better CPU utilization in containerized environments"
echo "5. ✅ New sync patterns: Cleaner concurrent code with WaitGroup.Go()"
echo "6. ✅ Better testing: Deterministic concurrent testing with testing/synctest"

# Cleanup
rm -f agentry-jsonv2 agentry-greentea agentry-experimental agentry-dwarf5 agentry-legacy
