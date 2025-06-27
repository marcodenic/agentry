# Agentry Bug Fix and Enhancement Plan

## Phase 1: Critical Bug Fixes (Immediate)

### 1.1 Fix Agent Spawning Error ✅ COMPLETED

**Issue**: `/spawn coder` fails with "tool names are reserved" error
**Root Cause**: Name conflict detection logic needs investigation
**Fix**:

- [x] Debug why "powershell" name is being used instead of "coder"
- [x] Fix agent name assignment in spawn command
- [x] Test manual spawn commands

### 1.2 Fix Token/Cost Tracking ✅ COMPLETED

**Issue**: Footer shows tokens: 0 and cost: $0.0000 despite API usage
**Root Cause**: Cost tracking not properly updating
**Fix**:

- [x] Verify Cost field is being populated by API calls
- [x] Ensure token counts accumulate across messages
- [x] Add debug logging to track cost updates
- [x] Fixed footer to show total across all agents
- [x] Added periodic refresh for live updates

### 1.3 Fix Spinner Issues ✅ COMPLETED

**Issue**: Spinners getting stuck and not clearing
**Root Cause**: Animation cleanup on completion
**Fix**:

- [x] Ensure thinking animation stops when tokens arrive
- [x] Clean up spinner state on agent stop
- [x] Fix spinner replacement logic
- [x] Improved spinner cleanup in tokenMsg handler

### 1.4 Fix Command Formatting ✅ COMPLETED

**Issue**: Commands lack proper spacing/grouping
**Root Cause**: No formatted command output system
**Fix**:

- [x] Add command grouping with proper spacing
- [x] Standardize command output formatting
- [x] Add visual separators for command sequences
- [x] Implemented formatSingleCommand and formatCommandGroup functions

## Phase 2: Orchestration Improvements

### 2.1 Auto-Delegation Logic

**Issue**: Agent 0 doesn't automatically spawn agents
**Root Cause**: Missing auto-delegation decision logic
**Fix**:

- [ ] Implement task analysis for auto-spawning
- [ ] Add decision tree for when to delegate
- [ ] Create agent type selection logic

### 2.2 Inter-Agent Communication

**Issue**: No agent-to-agent communication
**Root Cause**: Missing coordination system
**Fix**:

- [ ] Implement agent message passing
- [ ] Add shared context/memory
- [ ] Create task progress reporting

## Phase 3: Test Framework

### 3.1 Basic Orchestration Tests

- [ ] Simple task delegation test
- [ ] Multi-agent collaboration test
- [ ] Team formation test
- [ ] Resource allocation test

### 3.2 Advanced Workflow Tests

- [ ] Project management workflow
- [ ] Cross-agent communication test
- [ ] Dynamic scaling test
- [ ] Quality assurance workflow

## Implementation Order

1. **Fix spawn error** (highest priority - blocks testing)
2. **Fix token/cost tracking** (visibility into API usage)
3. **Fix spinner issues** (UX improvement)
4. **Improve command formatting** (readability)
5. **Add auto-delegation** (core orchestration feature)
6. **Add inter-agent communication** (advanced orchestration)
7. **Build test framework** (validation and demos)

## Success Criteria

- [ ] Manual `/spawn coder` works without errors
- [ ] Token counts and costs update correctly in footer
- [ ] Spinners clear properly when agents stop/complete
- [ ] Commands have proper spacing and visual grouping
- [ ] Agent 0 automatically spawns agents for appropriate tasks
- [ ] Multiple agents can work together on complex tasks
- [ ] Full test suite demonstrates all orchestration capabilities
