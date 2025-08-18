# 🚨 CRITICAL INSTRUCTIONS - READ FIRST 🚨

## ⚠️ MANDATORY TESTING PROTOCOL ⚠️

### 🔒 SANDBOX TESTING ONLY
**ALL AGENT COORDINATION TESTING MUST BE DONE IN:**
```
/tmp/agentry-ai-sandbox
```
**NEVER EVER test agents in the working directory `/home/marco/Documents/GitHub/agentry`**
- Agent testing can create/modify/delete files
- Working directory contains source code that must be protected
- Always copy required files to sandbox before testing

### 🔑 API KEY LOCATION
**API keys are ALWAYS in:**
```
/home/marco/Documents/GitHub/agentry/.env.local
```
**NEVER assume API keys are missing**
- The `.env.local` file contains `OPENAI_API_KEY`
- Copy `.env.local` to sandbox for testing
- Source the environment file: `source .env.local`

### 📋 MANDATORY SANDBOX SETUP
```bash
# 1. Create sandbox
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
echo "API Key set: ${OPENAI_API_KEY:0:10}..."
ls -la

# 5. Run tests safely
./agentry.exe "test command"
```

### 🛡️ SAFETY RULES
1. **ALWAYS** verify you're in `/tmp/agentry-ai-sandbox` before running agents
2. **ALWAYS** copy `.env.local` to sandbox
3. **ALWAYS** source `.env.local` before testing
4. **NEVER** run agent tests in working directory
5. **NEVER** assume API keys are missing - they're in `.env.local`

---

## 🎯 CURRENT PROJECT STATUS

### Agent 0 Coordination Testing
- ✅ Tool restriction implemented and working
- 🔄 Testing delegation workflow in sandbox
- 📍 Location: `/tmp/agentry-ai-sandbox` (MANDATORY)

### Next Steps
1. Set up sandbox with `.env.local`
2. Test Agent 0 delegation workflow
3. Validate coordination tools work
4. Document results

---

**⚠️ VIOLATION OF THESE RULES RISKS DATA LOSS AND PROJECT DAMAGE ⚠️**
