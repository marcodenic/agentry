# Agentry Product Notes

## New Feature: Agent TODO List Tool

### Motivation
Agents need a way to keep track of tasks, decisions, and open threads during multi-step orchestration. A shared or per-agent TODO list tool will:
- Help agents plan, coordinate, and remember what needs to be done
- Allow agents to add, update, delete, and refer to TODO items
- Enable persistent memory of open tasks across agent runs
- Reduce cognitive load by externalizing task memory

### Proposed Tool API
```json
{
  "todo_add": {
    "description": "Add a new TODO item",
    "params": {
      "title": "string (required)",
      "description": "string (optional)",
      "priority": "low|medium|high (optional, default: medium)",
      "tags": "string[] (optional)",
      "agent_id": "string (optional, defaults to current agent)"
    }
  },
  "todo_list": {
    "description": "List TODO items with filtering",
    "params": {
      "agent_id": "string (optional, filter by agent)",
      "status": "pending|done|all (optional, default: pending)",
      "tags": "string[] (optional, filter by tags)",
      "limit": "number (optional, default: 10)"
    }
  },
  "todo_update": {
    "description": "Update a TODO item",
    "params": {
      "id": "string (required)",
      "status": "pending|done (optional)",
      "title": "string (optional)",
      "description": "string (optional)",
      "priority": "low|medium|high (optional)"
    }
  },
  "todo_delete": {
    "description": "Remove a TODO item",
    "params": {
      "id": "string (required)"
    }
  },
  "todo_get": {
    "description": "Get details for a specific TODO item",
    "params": {
      "id": "string (required)"
    }
  }
}
```

### Storage Design
- Store in `memstore` with namespace `todo:project:<project_path>`
- Each TODO has: id, title, description, priority, tags, agent_id, status, created_at, updated_at
- TODOs persist across agent runs and can be shared between agents
- Auto-cleanup of old completed TODOs (configurable TTL)

### Example Usage
```javascript
// Coder agent planning work
todo_add({
  title: "Refactor health endpoint",
  description: "Extract health check logic for better testability",
  priority: "high",
  tags: ["refactor", "health"]
})

// Planner reviewing status
todo_list({status: "pending", limit: 5})

// Agent completing work  
todo_update({id: "todo_123", status: "done"})
```

---

## Context Window Management & Token Budgeting

### Critical Problem Identified
Recent agent runs hit Anthropic's 30k token/minute rate limit due to **excessive context injection**. Analysis of current context injection:

**Token Breakdown (BEFORE agent's role prompt + task):**
- Hardcoded project details: ~200 tokens
- Workspace activity (3 events): ~500-1000 tokens  
- Team coordination history (3 events): ~500-1000 tokens
- Intelligence guidelines: ~400 tokens
- Role-specific instructions: ~800 tokens
- **TOTAL: 2,400-3,400 tokens of overhead**

### Problems with Current Approach
1. ✅ **Hardcoded assumptions**: ~~"You are in a Go project called 'agentry'" - not dynamic~~ **FIXED**: Now uses dynamic project detection via `buildRootFileTree()`
2. **No token budgeting**: No awareness of model's context window limits
3. **No relevance scoring**: All context treated equally important
4. **Additive only**: No way to truncate or prioritize context
5. **Rate limit breach**: Can exceed API limits before real work begins

### Solution Architecture
Implement intelligent context management:

```
Provider → Budget → Assembler → Agent
   ↓         ↓         ↓
Context    Token    Ranked
Sources    Limits   Context
```

**Context Budget by Model:**
- GPT-4o-mini: ~75% of 128k = 96k tokens (aggressive budget)
- Claude-3.5-Sonnet: ~75% of 200k = 150k tokens  
- Fallback: 75% of 4k = 3k tokens for older models

**Context Pack Providers:**
- Project structure (dynamic discovery)
- Recent activity (last N events, scored by relevance)
- Tool usage patterns
- Conversation history (summarized if needed)
- Error context (if recent failures)

**Relevance Scoring:**
- Task keywords match workspace files: +10
- Recent agent activity on similar tasks: +5  
- Coordination events involving current agent: +3
- Generic workspace info: +1

### Implementation Priority
1. **Remove hardcoded context** from `buildContextualInput()` 
2. **Add token counting** to context injection
3. **Implement context packs** with scoring
4. **Add budget enforcement** per model type
5. **Test with realistic workloads** across model providers

---

## Next Steps
- Implement the TODO tool for all agents
- Refactor context injection to use packs, profiles, and budgeting
- Test with both OpenAI and Anthropic models, adjusting context size as needed
- Document context management and TODO tool in CONTEXT.md and PRODUCT.md
