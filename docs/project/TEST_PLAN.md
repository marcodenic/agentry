# (Archived) Agentry Bandwidth Monitor Test

> **Status:** Archived scenario. The bandwidth monitor project is no longer an active objective; keep these notes only if you revisit multi-agent experimentation.

> **⚠️ CRITICAL:** Before running ANY tests, read [./CRITICAL_INSTRUCTIONS.md](./CRITICAL_INSTRUCTIONS.md) for mandatory safety protocols and sandbox setup requirements.

**Version**: 3.0  
**Last Updated**: July 3, 2025  
**Focus**: Terminal-based bandwidth monitor development with multi-agent collaboration

---

## 🎯 **OBJECTIVE**

Test Agentry's ability to build a complete, working application through multi-agent collaboration. The target is a Go-based terminal bandwidth monitor that displays real-time upload/download statistics with color-coded charts.

### **Target Application: Terminal Bandwidth Monitor**
- **Tech Stack**: Go with terminal UI libraries
- **Core Features**: Real-time bandwidth monitoring, color-coded charts (green=download, red=upload, yellow=overlap), historical data
- **Success Criteria**: Compiles, runs, monitors bandwidth correctly, handles errors gracefully

### **Agent Team**
- **Agent 0**: System orchestrator and coordinator
- **Coder agents**: Implementation specialists 
- **Tester agents**: Testing and validation

## 🏗️ **SANDBOX SETUP**

**ALL TESTING MUST BE PERFORMED IN:**
```bash
/tmp/agentry-ai-sandbox
```

**Setup Protocol:**
```bash
# 1. Create and enter sandbox
mkdir -p /tmp/agentry-ai-sandbox
cd /tmp/agentry-ai-sandbox

# 2. Copy required files
cp /home/marco/Documents/GitHub/agentry/agentry.exe .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp -r /home/marco/Documents/GitHub/agentry/templates .

# 3. Source environment
source .env.local

# 4. Verify setup
./agentry.exe --version
```

## 🎯 **BANDWIDTH MONITOR DEVELOPMENT TEST**

### **Main Test: Complete Application Development**
**Duration**: 30-45 minutes  
**Objective**: Multi-agent team builds working bandwidth monitor from scratch

```bash
./agentry.exe chat << 'EOF'
PROJECT: Terminal Bandwidth Monitor in Go

REQUIREMENTS:
1. Monitor network interface bandwidth (upload/download rates)
2. Display real-time terminal-based charts
3. Use color coding: Green=download, Red=upload, Yellow=overlap
4. Show current rates and historical data (last 5-10 minutes)
5. Handle multiple network interfaces
6. Graceful error handling for network issues
7. Clean shutdown with Ctrl+C

TECHNICAL REQUIREMENTS:
- Go programming language
- Cross-platform (Linux/macOS/Windows) 
- Use appropriate terminal UI library (termui, tview, etc.)
- Efficient memory usage for historical data
- Configurable refresh rate (default 1 second)

DELIVERABLES:
1. Working Go application that compiles and runs
2. Clean, well-structured code with proper error handling
3. Terminal interface that displays bandwidth data clearly
4. Basic tests to validate core functionality

COLLABORATION REQUIREMENTS:
- Use collaborative tools for coordination between agents
- Agent 0 coordinates overall project
- Coder agents implement different components
- Tester agents validate functionality and run tests
- Agents communicate about dependencies and integration

VALIDATION:
- Application compiles successfully: go build
- Application runs without crashing
- Displays real bandwidth data from system
- Charts update in real-time
- Error handling works (simulated network issues)
- Code passes basic Go formatting and vet checks
EOF
```

### **Success Metrics**
- ✅ **Build Success**: `go build` completes without errors
- ✅ **Functionality**: Application monitors and displays bandwidth correctly
- ✅ **Visual Quality**: Clean terminal interface with proper colors
- ✅ **Collaboration**: 5+ collaborative tool calls between agents
- ✅ **Code Quality**: Passes `go fmt` and `go vet`
- ✅ **Error Handling**: Graceful handling of network interface issues

## 🔄 **ITERATIVE IMPROVEMENT CYCLE**

### **Bug Fix & Refinement Process**
If the initial build has issues, continue with iterative improvements:

```bash
./agentry.exe chat << 'EOF'
PHASE 2: Debug and Improve Bandwidth Monitor

SITUATION: Review the current state of the bandwidth monitor application

TASKS:
1. Test the current application and identify any issues
2. Fix compilation errors if any exist
3. Improve functionality that isn't working correctly
4. Enhance user experience and visual display
5. Optimize performance and resource usage

COLLABORATION:
- Tester agents identify and report specific issues
- Coder agents implement fixes for identified problems
- Agent 0 coordinates priorities and validates improvements
- Continue until application works reliably

VALIDATION:
- Each iteration should improve the application
- Document what was fixed and what still needs work
- Final goal: stable, working bandwidth monitor
EOF
```

## 📊 **VALIDATION METRICS**

### **Essential Success Criteria**
```bash
# Test in sandbox after development
cd /tmp/agentry-ai-sandbox/bandwidth-monitor

# 1. Compilation test
go build . && echo "✅ BUILD SUCCESS" || echo "❌ BUILD FAILED"

# 2. Basic functionality test  
timeout 10s ./bandwidth-monitor && echo "✅ RUNS OK" || echo "❌ RUNTIME ISSUES"

# 3. Code quality checks
go fmt ./... && echo "✅ FORMAT OK" || echo "❌ FORMAT ISSUES"
go vet ./... && echo "✅ VET CLEAN" || echo "❌ VET WARNINGS"
```

### **Collaboration Metrics**
- **Tool Usage**: Count collaborative tool calls in logs
- **Agent Communication**: Direct messages between agents
- **Problem Resolution**: Issues identified and fixed through teamwork

---

**🎯 This simplified test plan focuses on one concrete goal: building a working bandwidth monitor through multi-agent collaboration. Success is measured by a functional application that compiles, runs, and monitors bandwidth correctly.**
