#!/usr/bin/env python3
import json
import requests
import os

# Test specifically tool calling with Responses API using different prompts

api_key = os.getenv('OPENAI_API_KEY')
if not api_key:
    print("OPENAI_API_KEY not set")
    exit(1)

headers = {
    'Content-Type': 'application/json',
    'Authorization': f'Bearer {api_key}'
}

# Test with strong tool prompting
print("=== Testing /v1/responses with strong tool prompting ===")
responses_tools_payload = {
    "model": "gpt-4o-mini",
    "input": [
        {
            "role": "user",
            "content": [{"type": "input_text", "text": "I need you to use the bash tool to run the uptime command. Please call the bash tool with command 'uptime'"}]
        }
    ],
    "tools": [
        {
            "type": "function",
            "name": "bash",
            "description": "Execute bash commands",
            "parameters": {
                "type": "object",
                "properties": {
                    "command": {
                        "type": "string",
                        "description": "Bash command to execute"
                    }
                },
                "required": ["command"]
            }
        }
    ],
    "tool_choice": "required"
}

try:
    resp = requests.post('https://api.openai.com/v1/responses',
                        headers=headers,
                        json=responses_tools_payload)
    print(f"Status: {resp.status_code}")
    
    if resp.status_code == 200:
        data = resp.json()
        print("Full response structure:")
        print(json.dumps(data, indent=2))
    else:
        print(f"Error response: {resp.text}")
except Exception as e:
    print(f"Error: {e}")
