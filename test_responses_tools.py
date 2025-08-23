#!/usr/bin/env python3
import json
import requests
import os

# Test the correct tool format for Responses API

api_key = os.getenv('OPENAI_API_KEY')
if not api_key:
    print("OPENAI_API_KEY not set")
    exit(1)

headers = {
    'Content-Type': 'application/json',
    'Authorization': f'Bearer {api_key}'
}

# Test with the correct Responses API format (flat structure)
print("=== Testing /v1/responses with tools (CORRECT FORMAT) ===")
responses_tools_payload = {
    "model": "gpt-3.5-turbo",
    "input": [
        {
            "role": "user",
            "content": [{"type": "input_text", "text": "What's the weather like in Paris?"}]
        }
    ],
    "tools": [
        {
            "type": "function",
            "name": "get_weather",
            "description": "Get weather information",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {"type": "string"}
                },
                "required": ["location"]
            }
        }
    ]
}

try:
    resp = requests.post('https://api.openai.com/v1/responses',
                        headers=headers,
                        json=responses_tools_payload)
    print(f"Status: {resp.status_code}")
    if resp.status_code == 200:
        response_data = resp.json()
        print(f"Response ID: {response_data.get('id', 'N/A')}")
        print(f"Status: {response_data.get('status', 'N/A')}")
        
        # Look for output array
        output = response_data.get('output', [])
        print(f"Output items: {len(output)}")
        
        for i, item in enumerate(output):
            print(f"  Item {i}: type={item.get('type')}, role={item.get('role', 'N/A')}")
            if item.get('type') == 'message':
                content = item.get('content', [])
                for j, c in enumerate(content):
                    print(f"    Content {j}: type={c.get('type')}")
                    if c.get('type') == 'function_call':
                        print(f"      Function call: {c.get('name')} with args: {c.get('arguments')}")
                        
    else:
        print(f"Error Response: {resp.text}")
except Exception as e:
    print(f"Error: {e}")
