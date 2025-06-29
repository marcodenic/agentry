# Agentry Test Scripts

## Structure

### `/tests/coordination/`
Contains all test scripts related to Agent 0 coordination and delegation:
- `test_agent_0_tool_restrictions.sh` - Validates Agent 0's tool restrictions
- `validate_agent_0.sh` - Simple validation of Agent 0 functionality
- Various orchestration and delegation test scripts

### `/tests/archive/`
Contains older test scripts that are less relevant but kept for reference.

## Root Level Tests

### Essential Test Scripts
- `test_basic.sh` - Basic functionality test
- `test_chat_mode.sh` - Chat mode testing
- `test_interactive_session.sh` - Interactive session testing
- `test_cli_interactive.sh` - CLI interactive testing
- `test_quick_context.sh` - Quick context testing

## Usage

Most tests can be run directly:
```bash
./test_basic.sh
./test_chat_mode.sh
./tests/coordination/validate_agent_0.sh
```

For Agent 0 coordination testing, see `AGENT_0_STATUS.md` for current status and next steps.
