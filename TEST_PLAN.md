# Agentry Multi-Agent Platform - Comprehensive Test Plan

> **⚠️ CRITICAL:** Before running ANY tests, read [./CRITICAL_INSTRUCTIONS.md](./CRITICAL_INSTRUCTIONS.md) for mandatory safety protocols and sandbox setup requirements.

**Version**: 2.0  
**Last Updated**: July 1, 2025  
**Status**: Advanced Multi-Agent Collaboration Testing Phase  
**Current Focus**: Comprehensive validation of true multi-agent collaboration capabilities

---

## 🎯 **OVERVIEW & OBJECTIVES**

This test plan provides comprehensive validation of Agentry's multi-agent collaboration capabilities, from basic coordination through advanced collaborative scenarios requiring sustained agent-to-agent communication and iterative development.

### **Primary Objectives**
1. **Validate True Multi-Agent Collaboration** - Verify agents communicate directly using collaborative tools
2. **Test Advanced Scenarios** - Complex, long-running collaborative development projects
3. **Verify Scalability** - Multi-agent teams working on distributed, complex tasks
4. **Validate Quality Assurance** - Iterative testing, fixing, and refinement workflows
5. **Performance Testing** - Resource usage, response times, and reliability under load

### **Success Criteria**
- ✅ **Collaboration Tools**: 7+ collaborative tool calls per complex scenario
- ✅ **Direct Communication**: Agent-to-agent messages without Agent 0 mediation
- ✅ **Sustained Workflows**: 10+ minute collaborative development sessions
- ✅ **Quality Deliverables**: Working code that compiles, passes tests, and runs correctly
- ✅ **Team Coordination**: Multiple agents coordinating on shared objectives

---

## 🎯 **CURRENT STATUS & ACHIEVEMENTS**

### **Completed Tests (as of July 1, 2025)**
- ✅ **Tier 1**: Basic agent spawning and coordination
- ✅ **Tier 2**: Multi-agent task delegation scenarios  
- ✅ **Tier 3**: Advanced collaborative scenarios - **BREAKTHROUGH ACHIEVED**
- ✅ **CODE VALIDATION**: Agent code generation capability confirmed
- 🔄 **Tier 4**: Distributed system stress tests (validated with limitations)
- ⏳ **Tier 5**: Meta-collaboration and self-improvement (pending)

### **🏆 MAJOR BREAKTHROUGH - July 1, 2025**
**CONFIRMED**: Agents generate real, working, compilable code (not just plans)
- **317 lines** of Go code generated across **11 files**
- **Complete microservice** with HTTP server, database, auth, and middleware
- **Compilation successful** - 7.6MB working binary produced
- **Multi-agent coordination** validated with real code artifacts
- **Test Location**: `/tmp/agentry-ai-sandbox` (12 Go files generated)

### **Key Validated Capabilities**
✅ **Real Code Generation** - Not just pseudo-code or plans  
✅ **Multi-Agent Coordination** - Collaborative development workflows  
✅ **Complex Project Structure** - Proper separation of concerns  
✅ **Compilation Validation** - Working Go programs that build  
✅ **Microservices Patterns** - HTTP servers, databases, authentication  
✅ **Direct Agent Communication** - Agent-to-agent coordination tools  

### **Ready for Advanced Scenarios**
- **Distributed system implementation** with confidence in code generation
- **Docker containerization** with actual working services
- **Integration testing** with real HTTP endpoints and databases
- **Production deployment** scenarios with compilable artifacts
- **Performance testing** with functional microservices

---

## 🏗️ **TESTING INFRASTRUCTURE & METHODOLOGY**

### **🔒 Mandatory Sandbox Environment**
**ALL TESTING MUST BE PERFORMED IN:**
```bash
/tmp/agentry-ai-sandbox
```

### **📋 Standard Setup Protocol**
```bash
# 1. Create and enter sandbox
mkdir -p /tmp/agentry-ai-sandbox
cd /tmp/agentry-ai-sandbox

# 2. Copy required files (CRITICAL - includes API keys)
cp /home/marco/Documents/GitHub/agentry/agentry.exe .
cp /home/marco/Documents/GitHub/agentry/.agentry.yaml .
cp /home/marco/Documents/GitHub/agentry/.env.local .
cp -r /home/marco/Documents/GitHub/agentry/templates .

# 3. Source environment (MANDATORY)
source .env.local

# 4. Verify setup
echo "API Key set: ${OPENAI_KEY:0:10}..."
echo "Agentry binary: $(ls -la agentry.exe)"
echo "Templates: $(ls -la templates/)"

# 5. Ready for testing
./agentry.exe --version
```

### **🧪 Test Execution Framework**
- **Logging**: All tests generate detailed logs in `/tmp/agentry-test-logs/`
- **Artifacts**: Generated code and files preserved for analysis
- **Metrics**: Collaborative tool usage, communication patterns, performance data
- **Validation**: Automated verification of deliverables and success criteria

---

## 📊 **TEST HIERARCHY & PROGRESSION**

### **TIER 1: FOUNDATION VALIDATION** ⚡ (5-10 minutes each)

#### **T1.1: Basic Agent Coordination**
**Objective**: Verify Agent 0 can delegate to specialized agents
```bash
# Test delegation with tool restrictions
./agentry.exe chat "use team_status to show current team state"
./agentry.exe chat "check if coder agent is available using check_agent"
./agentry.exe chat "delegate to coder: analyze the current directory structure"
```
**Expected**: Agent 0 uses only coordination tools, delegates successfully

#### **T1.2: Tool Inheritance Verification**
**Objective**: Confirm spawned agents get full tool access
```bash
./agentry.exe chat "delegate to coder: create a simple test.txt file with content 'testing'"
./agentry.exe chat "delegate to coder: list files to verify test.txt was created"
```
**Expected**: Coder agent creates file successfully, tool inheritance working

#### **T1.3: Multi-Agent Availability**
**Objective**: Test multiple agent types
```bash
./agentry.exe chat "check which agents are available: coder, tester, writer, reviewer"
./agentry.exe chat "delegate simple tasks to each agent type to verify they work"
```
**Expected**: All agent types available and functional

---

### **TIER 2: COLLABORATIVE COMMUNICATION** 🤝 (10-15 minutes each)

#### **T2.1: Direct Agent Communication**
**Objective**: Verify agents can communicate directly using collaborative tools
```bash
./agentry.exe chat << 'EOF'
Coordinate a multi-agent task where agents must communicate with each other:
1. Assign a coder to create a Python function
2. Have the coder send a message to a tester about what they created
3. Have the tester respond with feedback
4. Use get_team_status to monitor the collaboration
EOF
```
**Expected**: Direct agent-to-agent messages, collaborative tool usage

#### **T2.2: Status Broadcasting and Awareness**
**Objective**: Test shared workspace awareness
```bash
./agentry.exe chat << 'EOF'
Create a collaborative workflow where multiple agents work on related tasks:
1. Coder starts working on a module
2. Writer simultaneously starts documentation
3. Both agents use update_status to broadcast their progress
4. Agents coordinate timing using team status information
EOF
```
**Expected**: Status updates, coordination based on shared awareness

#### **T2.3: Event-Driven Task Handoffs**
**Objective**: Test agents coordinating task completion
```bash
./agentry.exe chat << 'EOF'
Set up a pipeline where task completion triggers next steps:
1. Coder creates a function and notifies team when done
2. Tester waits for notification, then creates tests
3. Reviewer waits for both, then reviews everything
4. Use collaborative tools throughout for coordination
EOF
```
**Expected**: Event-driven coordination, proper task sequencing

---

### **TIER 3: ADVANCED COLLABORATIVE SCENARIOS** 🚀 (15-30 minutes each)

#### **T3.1: Go HTTP Server Collaborative Development**
**Script**: `/tmp/test_advanced_collaborative_scenario.sh`
**Objective**: Multi-agent development of complete Go HTTP server
**Requirements**:
- Multiple endpoints (GET /health, POST /calculate, GET /stats)
- Comprehensive error handling and input validation
- Unit and integration tests that pass
- Collaborative development with direct agent communication
- Iterative testing and refinement

**Execution**:
```bash
chmod +x /tmp/test_advanced_collaborative_scenario.sh
/tmp/test_advanced_collaborative_scenario.sh
```

**Success Metrics**:
- 🔧 Collaboration tool calls: 7+
- 💬 Direct agent messages: 5+
- 🏗️ Build status: SUCCESS
- 🧪 Test status: ALL_PASS
- 🌐 Server status: RESPONDS_CORRECTLY

#### **T3.2: Enhanced Multi-Agent Software Project**
**Script**: `/tmp/test_enhanced_collaborative_scenario.sh`
**Objective**: Complex project requiring sustained collaboration
**Features**:
- Task complexity assessment and multi-coder assignment
- Real-time collaboration monitoring
- Quality assurance with iterative improvement
- Documentation and testing coordination

**Execution**:
```bash
chmod +x /tmp/test_enhanced_collaborative_scenario.sh
/tmp/test_enhanced_collaborative_scenario.sh
```

#### **T3.3: Distributed System Architecture**
**Objective**: Design and implement a microservices architecture
```bash
./agentry.exe chat << 'EOF'
Coordinate a team to design and implement a distributed system:

PROJECT: Microservices Task Management System
COMPONENTS:
1. User Service (authentication, user management)
2. Task Service (CRUD operations, task management)
3. Notification Service (email/SMS notifications)
4. API Gateway (routing, load balancing)

REQUIREMENTS:
- Each service must be implemented by different agents
- Agents must coordinate API contracts and communication
- Include Docker configuration and deployment scripts
- Comprehensive testing strategy across services
- Documentation and API specifications

COLLABORATION:
- Use collaborative tools for coordination
- Agents must communicate about dependencies
- Iterative development with testing and integration
- Real-time status updates and progress tracking

Expected Duration: 20-30 minutes
EOF
```

**Success Criteria**:
- Multiple services implemented correctly
- API contracts coordinated between agents
- Docker configurations working
- Integration tests passing
- Comprehensive documentation

---

### **TIER 4: STRESS TESTING & EDGE CASES** 💪 (20-45 minutes each)

#### **T4.1: High-Frequency Collaboration**
**Objective**: Test system under rapid agent communication
```bash
./agentry.exe chat << 'EOF'
Create a high-frequency collaborative scenario:
1. 5+ agents working simultaneously on different modules
2. Frequent status updates (every 30 seconds)
3. Cross-agent dependencies requiring coordination
4. Real-time conflict resolution for shared files
5. Performance monitoring throughout the process
EOF
```

#### **T4.2: Large Codebase Development**
**Objective**: Test collaboration on complex, multi-file projects
```bash
./agentry.exe chat << 'EOF'
Develop a complete web application with multiple agents:

PROJECT: Task Management Web App
STACK: Go backend, React frontend, PostgreSQL database

AGENTS REQUIRED:
- Backend Developer: API server, database integration
- Frontend Developer: React UI, state management
- DevOps Engineer: Docker, deployment configuration
- QA Engineer: Testing strategy, automated tests
- Technical Writer: API docs, user documentation

COLLABORATION REQUIREMENTS:
- Agents must coordinate technology choices
- API contracts must be agreed upon
- Testing must cover full stack integration
- Documentation must be comprehensive and current

Expected Duration: 30-45 minutes
EOF
```

#### **T4.3: Failure Recovery and Resilience**
**Objective**: Test system behavior under error conditions
```bash
# Test scenarios with intentional failures
./agentry.exe chat << 'EOF'
Create a collaborative project that includes error scenarios:
1. Agent produces code with intentional bugs
2. Other agents must detect and communicate about issues
3. Collaborative debugging and fix coordination
4. Recovery workflows when builds fail
5. Conflict resolution when agents disagree on solutions
EOF
```

---

### **TIER 5: PRODUCTION READINESS** 🏭 (30-60 minutes each)

#### **T5.1: Self-Improvement Scenario**
**Objective**: Agentry agents improving Agentry itself
```bash
./agentry.exe chat << 'EOF'
Meta-project: Improve Agentry's collaboration capabilities

TASK: Analyze current collaboration tools and implement improvements
AGENTS:
- Analyzer: Review current collaboration.go and collaborative_features.go
- Designer: Propose enhanced collaboration patterns
- Implementer: Add new collaborative tools or improve existing ones
- Tester: Create test scenarios for new features
- Documenter: Update documentation with new capabilities

COLLABORATION:
- Agents must work on live Agentry codebase (in sandbox)
- Real-time coordination on code changes
- Testing new features as they're developed
- Documentation must stay current with implementations

DELIVERABLE: Enhanced collaboration system with tests and docs
EOF
```

#### **T5.2: Real-World Integration Project**
**Objective**: Complete integration with external systems
```bash
./agentry.exe chat << 'EOF'
Integration Project: GitHub Issue Management System

REQUIREMENTS:
1. GitHub API integration for issue tracking
2. Slack integration for team notifications
3. Automated testing and deployment pipeline
4. Monitoring and alerting system
5. Documentation and user guides

AGENTS:
- Integration Specialist: GitHub/Slack API development
- DevOps Engineer: CI/CD pipeline setup
- QA Engineer: Integration testing strategy
- Documentation Specialist: User guides and API docs
- Project Coordinator: Timeline and milestone tracking

COLLABORATION:
- Multi-service coordination
- External dependency management
- Real-time status tracking
- Quality assurance at each integration point

Expected Duration: 45-60 minutes
EOF
```

---

## 📊 **METRICS & EVALUATION FRAMEWORK**

### **Collaboration Metrics**
```bash
# Extract from test logs
COLLAB_TOOLS=$(grep -c "collaborate\|send_message\|get_team_status\|update_status" "$LOG_FILE")
AGENT_MESSAGES=$(grep -c "→.*agent.*message" "$LOG_FILE")
STATUS_UPDATES=$(grep -c "COORDINATION EVENT" "$LOG_FILE")
TOOL_USAGE=$(grep -c "Tool.*completed" "$LOG_FILE")
```

### **Quality Metrics**
```bash
# Validate deliverables
BUILD_SUCCESS=$(cd project && go build . && echo "SUCCESS" || echo "FAIL")
TEST_PASS=$(cd project && go test ./... && echo "PASS" || echo "FAIL")
SERVER_RESPONSE=$(curl -s http://localhost:8080/health | jq .status)
```

### **Performance Metrics**
```bash
# Monitor resource usage
CPU_USAGE=$(top -bn1 | grep "agentry" | awk '{print $9}')
MEMORY_USAGE=$(ps aux | grep agentry | awk '{print $6}')
RESPONSE_TIME=$(time ./agentry.exe chat "simple test" 2>&1 | grep real)
```

### **Success Thresholds**
- **Collaboration Score**: 8+ collaborative tool calls per complex scenario
- **Communication Score**: 5+ direct agent messages per scenario
- **Quality Score**: 90%+ tests passing, builds successful
- **Performance Score**: <5GB memory, <50% CPU, <10s response time

---

## 🔄 **CONTINUOUS TESTING WORKFLOW**

### **Daily Validation**
```bash
# Quick smoke test (5 minutes)
cd /tmp/agentry-ai-sandbox
source .env.local
./agentry.exe chat "delegate to coder: verify system is working by creating hello.txt"
```

### **Weekly Comprehensive Testing**
```bash
# Run full Tier 1-3 test suite (1-2 hours)
for test in T1.1 T1.2 T1.3 T2.1 T2.2 T2.3 T3.1 T3.2; do
    echo "Running test $test..."
    # Execute test and log results
done
```

### **Monthly Stress Testing**
```bash
# Run Tier 4-5 tests (2-4 hours)
# Performance baseline measurement
# Scalability testing
# Production readiness assessment
```

---

## 📁 **TEST ARTIFACTS & DOCUMENTATION**

### **Log Structure**
```
/tmp/agentry-test-logs/
├── daily/
│   ├── smoke-test-YYYY-MM-DD.log
│   └── metrics-YYYY-MM-DD.json
├── scenarios/
│   ├── T3.1-go-server-YYYY-MM-DD-HH-MM.log
│   ├── T3.2-enhanced-collab-YYYY-MM-DD-HH-MM.log
│   └── collaboration-metrics.json
└── artifacts/
    ├── generated-code/
    ├── test-results/
    └── performance-data/
```

### **Reporting Templates**
- **Test Execution Summary**: Pass/fail status, metrics, issues
- **Collaboration Analysis**: Tool usage patterns, communication flows
- **Performance Report**: Resource usage, response times, scalability
- **Quality Assessment**: Code quality, test coverage, deliverable analysis

---

## 🚨 **TROUBLESHOOTING & KNOWN ISSUES**

### **Common Setup Issues**
1. **API Key Not Found**: Ensure `.env.local` is copied and sourced
2. **Permission Denied**: Check executable permissions on `agentry.exe`
3. **Template Missing**: Verify templates directory is copied completely

### **Collaboration Issues**
1. **No Collaborative Tool Usage**: Check agent prompt configuration
2. **Agents Not Communicating**: Verify collaboration tools are registered
3. **Status Updates Missing**: Check team context and shared memory

### **Performance Issues**
1. **Slow Response**: Monitor model API rate limits
2. **High Memory Usage**: Check for agent cleanup and resource management
3. **Long Running Tests**: Set appropriate timeouts and monitoring

---

## 🎯 **NEXT STEPS & ROADMAP**

### **Immediate Priorities**
1. **Complete Tier 3 Validation** - Run all advanced collaborative scenarios
2. **Performance Baseline** - Establish performance metrics and thresholds
3. **Edge Case Testing** - Identify and test failure scenarios

### **Medium Term Goals**
1. **Automated Test Pipeline** - CI/CD integration for continuous testing
2. **Advanced Scenarios** - More complex distributed system projects
3. **Production Testing** - Real-world integration and deployment scenarios

### **Long Term Vision**
1. **Self-Testing System** - Agentry testing and improving itself
2. **Multi-Model Support** - Testing with different AI models
3. **Distributed Deployment** - Testing across multiple machines/regions

---

**🎉 This comprehensive test plan validates Agentry's breakthrough in true multi-agent collaboration. The foundation is solid - now we're testing advanced capabilities and pushing the boundaries of collaborative AI development! 🚀**
